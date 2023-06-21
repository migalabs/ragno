package peerDiscoverer

import (
	"github.com/cortze/ragno/pkg/csv"
	"github.com/cortze/ragno/pkg/modules"
)

type CsvPeerDiscoverer struct {
	sendingChan chan<- *modules.ELNode
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

func (c *CsvPeerDiscoverer) sendNodes(node *modules.ELNode) {
	c.sendingChan <- node
}

func (c *CsvPeerDiscoverer) ParseCsvToNodeInfo(lines [][]string) ([]*modules.ELNode, error) {
	// remove the header
	lines = lines[1:]

	// create the list of ELNodeInfo
	enrs := make([]*modules.ELNode, 0, len(lines)-1)

	// parse the file
	for _, line := range lines {
		// create the modules.ELNode
		elNodeInfo := new(modules.ELNode)
		elNodeInfo.Enode = modules.ParseStringToEnr(line[csv.ENR])
		elNodeInfo.Enr = line[csv.ENR]
		elNodeInfo.FirstTimeSeen = line[csv.FIRST_SEEN]
		elNodeInfo.LastTimeSeen = line[csv.LAST_SEEN]
		// add the struct to the list
		enrs = append(enrs, elNodeInfo)
	}
	return enrs, nil
}
