package models

import (
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type EnodeSet struct {
	m    sync.RWMutex
	list map[string]*ENR
}

func NewEnodeSet() *EnodeSet {
	return &EnodeSet{
		list: make(map[string]*ENR),
	}
}

func (s *EnodeSet) AddNode(n *ENR) error {
	if !n.IsValid() {
		return errors.New(fmt.Sprintf("attempt to persist non-valid node %+v", n))
	}
	s.m.Lock()
	defer s.m.Unlock()
	s.list[n.ID.String()] = n
	return nil
}

func (s *EnodeSet) GetENRs() []*ENR {
	enrs := make([]*ENR, s.Len())
	s.m.RLock()
	defer s.m.RUnlock()
	cnt := 0
	for _, enr := range s.list {
		enrs[cnt] = enr
		cnt++
	}
	return enrs
}

func (s *EnodeSet) RowComposer(rawRow []interface{}) []string {
	row := make([]string, len(rawRow), len(rawRow))
	for idx, item := range rawRow {
		switch item.(type) {
		case float64:
			row[idx] = fmt.Sprintf("%.6f", item.(float64))
		case int64:
			row[idx] = fmt.Sprintf("%d", item.(int64))
		case string:
			row[idx] = item.(string)
		case time.Duration:
			newItem := item.(time.Duration)
			row[idx] = fmt.Sprintf("%.6f", float64(newItem.Nanoseconds()))
		case DiscoveryType:
			t := item.(DiscoveryType)
			row[idx] = t.String()
		default:
			row[idx] = fmt.Sprint(item)
		}
	}
	return row
}

func (s *EnodeSet) generateRows() [][]interface{} {
	rows := make([][]interface{}, 0)
	// combine the arrays into rows for the csv
	s.m.RLock()
	defer s.m.RUnlock()
	for _, enr := range s.list {
		row := enr.ComposeCSVItems()
		rows = append(rows, row)
	}
	return rows
}

func (s *EnodeSet) PeerRows() [][]interface{} {
	return s.generateRows()
}

func (s *EnodeSet) Len() int {
	s.m.RLock()
	defer s.m.RUnlock()
	return len(s.list)
}
