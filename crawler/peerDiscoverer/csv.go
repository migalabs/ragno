package peerDiscoverer

import (
	"github.com/cortze/ragno/pkg/csv"
	"github.com/cortze/ragno/pkg/spec"
	"github.com/cortze/ragno/pkg/utils"
)

type CsvPeerDiscoverer struct {
	sendingChan chan<- *spec.ELNode
	csvImporter *csv.CSVImporter
}

func NewCSVPeerDiscoverer(conf PeerDiscovererConf) (PeerDiscoverer, error) {
	csvImporter, err := csv.NewCsvImporter(conf.File)
	if err != nil {
		return nil, err
	}

	disc := &CsvPeerDiscoverer{
		sendingChan: conf.SendingChan,
		csvImporter: csvImporter,
	}
	return disc, nil
}

func (c *CsvPeerDiscoverer) Run() error {
	// get all the lines from the CSV
	lines, err := c.csvImporter.Items()
	if err != nil {
		return err
	}
	peers, err := c.ParseCsvToNodeInfo(lines)
	if err != nil {
		return err
	}

	// send the peers to the sending channel
	for _, peer := range peers {
		c.sendNodes(peer)
	}

	return nil
}

func (c *CsvPeerDiscoverer) sendNodes(node *spec.ELNode) {
	c.sendingChan <- node
}

func (c *CsvPeerDiscoverer) ParseCsvToNodeInfo(lines [][]string) ([]*spec.ELNode, error) {
	// remove the header
	lines = lines[1:]

	// create the list of ELNodeInfo
	enrs := make([]*spec.ELNode, 0, len(lines)-1)

	// parse the file
	for _, line := range lines {
		// create the spec.ELNode
		elNodeInfo := new(spec.ELNode)
		elNodeInfo.Enode = utils.ParseStringToEnr(line[csv.ENR])
		elNodeInfo.Enr = line[csv.ENR]
		elNodeInfo.FirstTimeSeen = line[csv.FIRST_SEEN]
		elNodeInfo.LastTimeSeen = line[csv.LAST_SEEN]
		// add the struct to the list
		enrs = append(enrs, elNodeInfo)
	}
	return enrs, nil
}
