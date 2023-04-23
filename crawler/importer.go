package crawler

import (
	"bufio"
	"os"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

// for now only supports list of enr so far
type CSVImporter struct {
	path string
	items []*enode.Node
}

func NewCsvImporter(p string) (*CSVImporter, error) {

	importer := &CSVImporter {
		path: p,
		items: make([]*enode.Node, 0), 
	} 

	// open file and read content
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	csvScanner := bufio.NewScanner(f)
	csvScanner.Split(bufio.ScanLines)	
	for csvScanner.Scan() {
		row := csvScanner.Text()
		enr := ParseStringToEnr(row) 
		importer.items = append(importer.items, enr)
	}
	return importer, nil 
}

func (i *CSVImporter) Items() []*enode.Node {
	return i.items 
}


