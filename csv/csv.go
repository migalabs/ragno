package csv

import (
	"encoding/csv"
	"os"
	"strings"

	"github.com/cortze/ragno/modules"
)

const (
	// csv columns
	NODE_ID = iota
	FIRST_SEEN
	LAST_SEEN
	IP
	TCP
	UDP
	SEQ
	PK
	ENR
)

// for now only supports list of enr so far
type CSVImporter struct {
	path string
	r    *csv.Reader
}

func NewCsvImporter(p string) (*CSVImporter, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close() // Close the file when done

	fileContent, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	return &CSVImporter{
		path: p,
		r:    csv.NewReader(strings.NewReader(string(fileContent))),
	}, nil
}

func (i *CSVImporter) ReadELNodes() ([]*modules.ELNode, error) {
	// get the lines of the file
	lines, err := i.items()
	if err != nil {
		return nil, err
	}

	// remove the header
	lines = lines[1:]

	// create the list of ELNodeInfo
	enrs := make([]*modules.ELNode, 0, len(lines))

	// parse the file
	for _, line := range lines {
		// create the modules.ELNode
		elNodeInfo := new(modules.ELNode)
		elNodeInfo.Enode = modules.ParseStringToEnr(line[ENR])
		elNodeInfo.Enr = line[ENR]
		elNodeInfo.FirstTimeSeen = line[FIRST_SEEN]
		elNodeInfo.LastTimeSeen = line[LAST_SEEN]
		// add the struct to the list
		enrs = append(enrs, elNodeInfo)
	}
	return enrs, nil
}

func (i *CSVImporter) items() ([][]string, error) {
	return i.r.ReadAll()
}

func (i *CSVImporter) nextLine() ([]string, error) {
	return i.r.Read()
}

func (i *CSVImporter) changeSeparator(sep rune) {
	i.r.Comma = sep
}

func (i *CSVImporter) changeCommentChar(c rune) {
	i.r.Comment = c
}
