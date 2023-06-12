package crawler

import (
	"context"

	models "github.com/cortze/ragno/pkg"

	"github.com/sirupsen/logrus"
)

func Connect(ctx *context.Context, nodeInfo *models.ELNodeInfo, host *Host, savingChan chan *models.ELNodeInfo) error {

	logrus.Info("connecting to: ", nodeInfo.Enr)
	hinfo := host.Connect(nodeInfo.Enode)
	if hinfo.Error != nil {
		logrus.Error("Node: ", nodeInfo.Enr, hinfo.Error)
	}
	logrus.Info("connected to: ", nodeInfo.Enr)
	nodeInfo.Hinfo = hinfo
	savingChan <- nodeInfo
	return hinfo.Error
}
