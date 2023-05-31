package crawler

import (
	"bufio"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

// for now only supports list of enr so far
type CSVImporter struct {
	path  string
	rows  [][]string
	items []*enode.Node
}

func NewCsvImporter(p string) (*CSVImporter, error) {
	importer := &CSVImporter{
		path:  p,
		rows:  make([][]string, 0),
		items: make([]*enode.Node, 0),
	}

	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close() // Close the file when done

	csvScanner := bufio.NewScanner(f)
	csvScanner.Split(bufio.ScanLines)
	csvScanner.Scan() // skip the header

	for csvScanner.Scan() {
		row := csvScanner.Text()
		cols := strings.Split(row, ",")
		if len(cols) > 0 {
			enr_str := cols[len(cols)-1]
			enr := ParseStringToEnr(enr_str)
			usefull_cols := []string{cols[1], cols[2], cols[8]} 
			importer.rows = append(importer.rows, usefull_cols)
			importer.items = append(importer.items, enr)
		}
	}

	return importer, nil
}

func (i *CSVImporter) Items() []*enode.Node {
	return i.items
}

func (i *CSVImporter) Infos() [][]string{
	return i.rows
}