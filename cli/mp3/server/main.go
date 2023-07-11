package main

import (
	"os"
	"time"

	"github.com/zonglinpeng/distributed_algorithms/cmd/mp3/server"
	"github.com/zonglinpeng/distributed_algorithms/lib/logger"
	sync "github.com/sasha-s/go-deadlock"
	log "github.com/sirupsen/logrus"
)

func main() {
	sync.Opts.DeadlockTimeout = time.Second * 1000
	logger.SetupLogger(log.StandardLogger())
	rootCMD := server.NewRootCMD()
	if err := rootCMD.Execute(); err != nil {
		log.Errorf("%v\n", err)
		os.Exit(1)
	}
}
