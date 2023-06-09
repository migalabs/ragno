package crawler

import (
	"encoding/csv"
	"os"
	"strings"
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

func (i *CSVImporter) Items() ([][]string, error) {
	return i.r.ReadAll()
}

func (i *CSVImporter) NextLine() ([]string, error) {
	return i.r.Read()
}

func (i *CSVImporter) ChangeSeparator(sep rune) {
	i.r.Comma = sep
}

func (i *CSVImporter) ChangeCommentChar(c rune) {
	i.r.Comment = c
}
