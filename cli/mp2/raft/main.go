package main

import (
	"os"
	"time"

	"github.com/zonglinpeng/distributed_algorithms/cmd/mp2"
	"github.com/zonglinpeng/distributed_algorithms/lib/logger"
	sync "github.com/sasha-s/go-deadlock"
	log "github.com/sirupsen/logrus"
)

func main() {
	sync.Opts.DeadlockTimeout = time.Second * 100
	logger.SetupLogger(log.StandardLogger())
	rootCMD := mp2.NewRootCMD()
	if err := rootCMD.Execute(); err != nil {
		log.Errorf("%v\n", err)
		os.Exit(1)
	}
}
