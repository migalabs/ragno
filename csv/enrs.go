package csvs

import "github.com/cortze/ragno/models"

func (c *CSV) ReadENRset() (*models.EnodeSet, error) {
	lines, err := c.items()
	if err != nil {
		return nil, err
	}
	lines = lines[1:] // remove the header

	enrSet := models.NewEnodeSet()
	for _, line := range lines {
		enr, err := models.NewENR(models.FromCSVline(line))
		if err != nil {
			return nil, err
		}
		enrSet.AddNode(enr)
	}
	return enrSet, nil
}
