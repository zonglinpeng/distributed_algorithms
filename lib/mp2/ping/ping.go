package ping

import (
	"strconv"
	"time"

	"github.com/zonglinpeng/distributed_algorithms/lib/mp2/group"
	log "github.com/sirupsen/logrus"
)

var (
	logger = log.WithField("src", "ping")
)

type Ping struct {
	g *group.Group
}

func New(g *group.Group) *Ping {
	return &Ping{
		g: g,
	}
}

func (p *Ping) Run() {
	numOfNodes := p.g.NumOfNodes()
	nidStr := p.g.NID()
	nid, _ := strconv.Atoi(nidStr)
	nextNID := (nid + 1) % numOfNodes
	nextNIDStr := strconv.Itoa(nextNID)

	p.g.Route("ping", func(ctx *group.RecvContext) {
		logger.Infof("RECV %s", ctx)
	})

	p.g.Start()

	for {
		p.g.Unicast(nextNIDStr, map[string]string{
			"path": "ping",
			"ping": nidStr,
		})
		time.Sleep(time.Second * 5)
	}
}
