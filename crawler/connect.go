package crawler

import (
	"context"

	"github.com/cortze/ragno/pkg/spec"
	"github.com/sirupsen/logrus"
)

func Connect(ctx *context.Context, nodeInfo *spec.ELNode, host *Host) {

	logrus.Info("connecting to: ", nodeInfo.Enr)
	nodeInfo.Hinfo = host.Connect(nodeInfo.Enode)
	if nodeInfo.Hinfo.Error != nil {
		logrus.Error("Node: ", nodeInfo.Enr, ": ", nodeInfo.Hinfo.Error)
	} else {
		logrus.Info("Node: ", nodeInfo.Enr, " connected")
	}
}
