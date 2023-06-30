package crawler

import (
	"os"
	"github.com/cortze/ragno/modules"
)

type CsvExporter struct {
	fileName string 
	f *os.File
	headers []string
}

func NewCsvExporter(f string, headers []string) (*CsvExporter, error) {
	// check if file exists
	csvF, err := os.Create(f)
	if err != nil {
		return nil, err
	}
	csve := &CsvExporter{
		fileName: f,
		f: csvF,
		headers: headers,
	}
	return csve, err
}

func (c *CsvExporter) composeRow(items []string) string {
	newRow := ""
	for _, i := range items {
		if newRow == "" {
			newRow = i
			continue
		}
		newRow = newRow + "," + i
	}
	return newRow+"\n"
}

func (c *CsvExporter) writeLine(row string) error {
	_, err := c.f.WriteString(row)
	return err
} 

func (c *CsvExporter) Export(peers []*modules.EthNode) error {
	// reset index in the file
	c.f.Seek(0,0)
	defer c.f.Sync()
	headerRow := c.composeRow(c.headers)
	err := c.writeLine(headerRow)
	if err != nil {
		return nil
	}
	for _, n := range peers {
		row := c.composeRow(n.ComposeCSVItems()) 
		err = c.writeLine(row)
		if err != nil {
			return nil
		}
	}
	return nil
}

func (c *CsvExporter) Close() {
	c.f.Close()
}


