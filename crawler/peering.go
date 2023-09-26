package crawler

import "context"

type Peering struct {
	ctx     context.Context
	workers int
	host    *Host
	nodeSet *nodeSet
}

type nodeSet struct {
}

func (s *nodeSet) updateFromDB() {

}

func (s *nodeSet) orderSet() {

}

// sorting
