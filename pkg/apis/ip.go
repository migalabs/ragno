package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/cortze/ragno/models"
)

const (
	defaultIpTTL   = 30 * 24 * time.Hour // 30 days
	ipChanBuffSize = 45                  // number of ips that can be buffered unto the channel
	ipBuffSize     = 8192                // number of ip queries that can be queued in the ipQueue
	ipApiEndpoint  = "http://ip-api.com/json/{__ip__}?fields=status,continent,continentCode,country,countryCode,region,regionName,city,zip,lat,lon,isp,org,as,asname,mobile,proxy,hosting,query"
	minIterTime    = 100 * time.Millisecond
)

var ErrTooManyRequests error = fmt.Errorf("error HTTP 429")

// DB Interface for PostgresDBService
type PostgresDBService interface {
	PersistIPInfo(models.IPInfo)
	GetIPInfo(string) (models.IPInfo, error)
	CheckIPRecords(string) (bool, bool, error)
	GetExpiredIPInfo() ([]string, error)
}

// PEER LOCALIZER
type IPLocator struct {
	ctx context.Context
	// Request channels
	locationRequest chan string

	// dbClient
	dbClient PostgresDBService

	ipQueue *ipQueue
	// control variables for IP-API request
	apiCalls *int32
}

func NewIPLocator(ctx context.Context, dbCli PostgresDBService) *IPLocator {
	calls := int32(0)
	return &IPLocator{
		ctx:             ctx,
		locationRequest: make(chan string, ipChanBuffSize),
		dbClient:        dbCli,
		apiCalls:        &calls,
		ipQueue:         newIpQueue(ipBuffSize),
	}
}

// Run the necessary routines to locate IPs
func (ipLoc *IPLocator) Run() {
	go ipLoc.queueRoutine()
	go ipLoc.locatorRoutine()
}

func (ipLoc *IPLocator) queueRoutine() {
	ticker := time.NewTicker(minIterTime)
	for {
		ip, err := ipLoc.ipQueue.readItem()
		if err == nil {
			// put the request in the Queue
			ipLoc.locationRequest <- ip
		}
		select {
		case <-ticker.C:
			ticker.Reset(minIterTime)

		case <-ipLoc.ctx.Done():
			return
		}
	}
}

// locatorRoutine is the main routine that will wait until an request to identify an IP arrives
// or if the routine gets canceled
func (ipLoc *IPLocator) locatorRoutine() {
	log.Info("IP locator routine started")

	var nextDelayRequest time.Duration
	for {
		select {
		case ipRequest := <-ipLoc.locationRequest:
			log.Trace("new request has been received for ip:", ipRequest)
		requestLoop:
			for {
				// since it didn't exist or did expire, request the ip
				// new API call needs to be done
				log.Tracef("making API call for %s", ipRequest)
				atomic.AddInt32(ipLoc.apiCalls, 1)
				respC := ipLoc.locateIp(ipRequest)
				select {
				case response := <-respC:
					nextDelayRequest = response.DelayTime
					log.WithFields(log.Fields{
						"delay":         nextDelayRequest,
						"attempts left": response.AttemptsLeft,
					}).Debug("got response from IP-API request ")
					// check if there is an error
					switch response.Err {
					case ErrTooManyRequests:
						// if the error reports that we tried too many calls on the API, sleep given time and try again
						log.Debug("call ", ipRequest, " -> error received: ",
							response.Err.Error(), "\nwaiting ", nextDelayRequest+(5*time.Second))
						ticker := time.NewTicker(nextDelayRequest + (5 * time.Second))
						select {
						case <-ticker.C:
							continue
						case <-ipLoc.ctx.Done():
							log.Info("context closure has been detecting, closing IpApi caller")
							return
						}
					case nil:
						// if the error is different from ErrTooManyRequests break loop and store the request
						log.Debugf("call %s-> api req success", ipRequest)
						// Upsert the IP into the db
						ipLoc.dbClient.PersistIPInfo(response.IPInfo)
						break requestLoop

					default:
						log.Debug("call ", ipRequest, " -> diff error received: ", response.Err.Error())
						break requestLoop

					}

				case <-ipLoc.ctx.Done():
					log.Info("context closure has been detecting, closing IpApi caller")
					return
				}
			}
			// check if there is any waiting time that we have to respect before next connection
			if nextDelayRequest != time.Duration(0) {
				log.Debug("number of allowed requests has been exceed, waiting ", nextDelayRequest+(2*time.Second))
				// set req delay to true, noone can make requests
				ticker := time.NewTicker(nextDelayRequest + (2 * time.Second))
				select {
				case <-ticker.C:
					continue
				case <-ipLoc.ctx.Done():
					log.Info("context closure has been detecting, closing IpApi caller")
					return
				}
			}

		// the context has been deleted, end go routine
		case <-ipLoc.ctx.Done():
			// close the channels
			close(ipLoc.locationRequest)
			return
		}
	}
}

// LocateIP is an externa request that any module could do to identify an IP
func (ipLoc *IPLocator) LocateIP(ip string) {
	// check first if IP is already in queue (to queue same ip)
	if ipLoc.ipQueue.ipExists(ip) {
		return
	}

	// Check if the IP is already in the cache
	exists, expired, err := ipLoc.dbClient.CheckIPRecords(ip)
	if err != nil {
		log.Error("unable to check if IP already exists -", err.Error()) // Should it be a Panic?
	}
	// if exists and it didn't expired, don't do anything
	if exists && !expired {
		return
	}

	// since it didn't exist or it is expired, locate it again
	ticker := time.NewTicker(1 * time.Second)
	// wait 1 sec because is the normal time to wait untill we can start querying again
	for {
		err := ipLoc.ipQueue.addItem(ip)
		if err == nil {
			break
		}
		<-ticker.C
		ticker.Reset(1 * time.Second)
		log.Debug("waiting to alocate a new IP request")
	}
	ticker.Stop()
}

func (ipLoc *IPLocator) Close() {
	log.Info("closing IP-API service")
	// close the context for ending up the routine

}

func (ipLoc *IPLocator) locateIp(ip string) chan models.IPInfoResponse {
	respC := make(chan models.IPInfoResponse)
	go callIpApi(ip, respC)
	return respC
}

// get location country and City from the multiaddress of the peer on the peerstore
func callIpApi(ip string, respC chan models.IPInfoResponse) {
	var IPInfoResponse models.IPInfoResponse
	IPInfoResponse.IPInfo, IPInfoResponse.DelayTime, IPInfoResponse.AttemptsLeft, IPInfoResponse.Err = CallIpApi(ip)
	respC <- IPInfoResponse
	// defer ^
}

func CallIpApi(ip string) (iPInfo models.IPInfo, delay time.Duration, attemptsLeft int, err error) {
	url := strings.Replace(ipApiEndpoint, "{__ip__}", ip, 1)

	// Make the IP-APi request
	resp, err := http.Get(url)
	if err != nil {
		err = errors.Wrap(err, "unable to locate IP"+ip)
		return
	}
	timeLeft, _ := strconv.Atoi(resp.Header["X-Ttl"][0])
	// check if the error that we are receiving means that we exeeded the request limit
	if resp.StatusCode == 429 {
		log.Debugf("limit of requests per minute has been exeeded, wait for next call %s secs",
			resp.Header["X-Ttl"][0])
		err = ErrTooManyRequests
		delay = time.Duration(timeLeft) * time.Second
		return
	}

	// Check the attempts left that we have to call the api
	attemptsLeft, _ = strconv.Atoi(resp.Header["X-Rl"][0])
	if attemptsLeft <= 0 {
		// if there are no more attempts left, check how much time is needed
		// until we can call it again, and set it as delayTime
		delay = time.Duration(timeLeft) * time.Second
	}

	// check if the response was success or not
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		err = errors.Wrap(err, "could not read response body")
		return
	}

	var apiMsg models.IPInfoMsg
	// Convert response body to struct
	err = json.Unmarshal(bodyBytes, &apiMsg)
	if err != nil {
		err = errors.Wrap(err, "could not unmarshall response")
		return
	}
	// Check if the status of the request has been succesful
	if apiMsg.Status != "success" {
		err = errors.New(fmt.Sprintf("status from ip different than success, resp header:\n %#v \n %+v", resp, apiMsg))
		return
	}

	iPInfo.ExpirationTime = time.Now().UTC().Add(defaultIpTTL)
	iPInfo.IPInfoMsg = apiMsg
	return
}

func newIpQueue(queueSize int) *ipQueue {
	return &ipQueue{
		queueSize: queueSize,
		ipList:    make([]string, 0, queueSize),
	}
}

var (
	ErrorQueueFull  = errors.New("queue is full")
	ErrorQueueEmpty = errors.New("queue is emtpy")
)

type ipQueue struct {
	sync.RWMutex
	queueSize int
	ipList    []string
}

func (q *ipQueue) ipExists(target string) bool {
	for _, ip := range q.ipList {
		if ip == target {
			return true
		}
	}
	return false
}

func (q *ipQueue) addItem(newItem string) error {
	q.Lock()
	defer q.Unlock()

	if q.len() >= q.queueSize {
		return ErrorQueueFull
	}

	if q.ipExists(newItem) {
		return nil
	}

	q.ipList = append(q.ipList, newItem)

	return nil
}

func (q *ipQueue) readItem() (string, error) {
	q.Lock()
	defer q.Unlock()

	var item string
	if q.len() <= 0 {
		return item, ErrorQueueEmpty
	}

	item = q.ipList[0]

	// remove after the item from the list
	q.ipList = append(q.ipList[:0], q.ipList[0+1:]...)

	return item, nil
}


func (q *ipQueue) len() int {
	return len(q.ipList)
}

func (q *ipQueue) Len() int {
	q.RLock()
	defer q.RUnlock()

	return q.len()
}
