package models

import (
	"fmt"
	"sync"

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
	s.list[n.ID] = n
	return nil
}

func (s *EnodeSet) PeerList() []*ENR {
	plist := make([]*ENR, 0, len(s.list))
	s.m.Lock()
	defer s.m.Unlock()
	for _, v := range s.list {
		plist = append(plist, v)
	}
	return plist
}

func (s *EnodeSet) Len() int {
	s.m.RLock()
	defer s.m.RUnlock()
	return len(s.list)
}
