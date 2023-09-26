package csvs

import (
	"encoding/csv"
	"os"
	"strings"
)

type RowComposer func([]interface{}) []string

// for now only supports list of enr so far
type CSV struct {
	file    string
	columns []string
	// exporter
	f *os.File
	// importer
	r *csv.Reader
}

func NewCsvExporter(file string, columns []string) (*CSV, error) {
	csvFile, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	return &CSV{
		file:    file,
		columns: columns,
		f:       csvFile,
	}, nil
}

func NewCsvImporter(file string) (*CSV, error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return nil, err
	} // Close the file when done

	fileContent, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return &CSV{
		file: file,
		r:    csv.NewReader(strings.NewReader(string(fileContent))),
	}, nil
}

// --- Exporter ----
func (c *CSV) composeRow(items []string) string {
	newRow := ""
	for _, i := range items {
		if newRow == "" {
			newRow = i
			continue
		}
		newRow = newRow + "," + i
	}
	return newRow + "\n"
}

func (c *CSV) writeLine(row string) error {
	_, err := c.f.WriteString(row)
	return err
}

func (c *CSV) Export(rows [][]interface{}, rowComposer RowComposer) error {
	// reset index in the file
	c.f.Seek(0, 0)
	defer c.f.Sync()
	headerRow := c.composeRow(c.columns)
	err := c.writeLine(headerRow)
	if err != nil {
		return nil
	}
	for _, row := range rows {
		row := c.composeRow(rowComposer(row))
		err = c.writeLine(row)
		if err != nil {
			return nil
		}
	}
	return nil
}

func (c *CSV) Close() error {
	return c.f.Close()
}

// --- Importer ---
func (i *CSV) items() ([][]string, error) {
	return i.r.ReadAll()
}

func (i *CSV) nextLine() ([]string, error) {
	return i.r.Read()
}

func (i *CSV) changeSeparator(sep rune) {
	i.r.Comma = sep
}

func (i *CSV) changeCommentChar(c rune) {
	i.r.Comment = c
}
