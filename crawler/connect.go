package crawler

import (
	"github.com/cortze/ragno/pkg/modules"
	"github.com/sirupsen/logrus"
)

func (c *Crawler) Connect(nodeInfo *modules.ELNode) {

	nodeInfo.Hinfo = c.host.Connect(nodeInfo.Enode)
	if nodeInfo.Hinfo.Error != nil {
		logrus.Trace("Node: ", nodeInfo.Enr, ": ", nodeInfo.Hinfo.Error)
	} else {
		logrus.Trace("Node: ", nodeInfo.Enr, " connected")
	}
}
