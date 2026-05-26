package gossip

import (
	"github.com/hashicorp/memberlist"
	"go.uber.org/zap"
)

type eventDelegate struct {
	logger  *zap.Logger
	onJoin  func(node *memberlist.Node)
	onLeave func(node *memberlist.Node)
}

func (d *eventDelegate) NotifyJoin(node *memberlist.Node) {
	d.logger.Info("node joined", zap.String("name", node.Name), zap.String("addr", node.Address()))
	if d.onJoin != nil {
		d.onJoin(node)
	}
}

func (d *eventDelegate) NotifyLeave(node *memberlist.Node) {
	d.logger.Info("node left", zap.String("name", node.Name), zap.String("addr", node.Address()))
	if d.onLeave != nil {
		d.onLeave(node)
	}
}

func (d *eventDelegate) NotifyUpdate(node *memberlist.Node) {
	d.logger.Debug("node updated", zap.String("name", node.Name))
}
