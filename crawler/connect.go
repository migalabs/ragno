package crawler

import (
	"context"

	models "github.com/cortze/ragno/pkg/models"

	"github.com/sirupsen/logrus"
)

func Connect(ctx *context.Context, nodeInfo *models.ELNodeInfo, host *Host) {

	logrus.Info("connecting to: ", nodeInfo.Enr)
	nodeInfo.Hinfo = host.Connect(nodeInfo.Enode)
	if nodeInfo.Hinfo.Error != nil {
		logrus.Error("Node: ", nodeInfo.Enr, ": ", nodeInfo.Hinfo.Error)
	} else {
		logrus.Info("Node: ", nodeInfo.Enr, " connected")
	}
}
