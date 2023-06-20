package crawler

import (
	"context"

	"github.com/cortze/ragno/pkg/spec"
	"github.com/sirupsen/logrus"
)

func Connect(ctx *context.Context, nodeInfo *spec.ELNode, host *Host) {

	nodeInfo.Hinfo = host.Connect(nodeInfo.Enode)
	if nodeInfo.Hinfo.Error != nil {
		logrus.Trace("Node: ", nodeInfo.Enr, ": ", nodeInfo.Hinfo.Error)
	} else {
		logrus.Trace("Node: ", nodeInfo.Enr, " connected")
	}
}
