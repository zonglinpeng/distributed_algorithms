package main

import (
	"os"

	"github.com/zonglinpeng/distributed_algorithms/cmd/mp0/server"
	"github.com/zonglinpeng/distributed_algorithms/lib/logger"
	log "github.com/sirupsen/logrus"
)

func main() {
	logger.SetupLogger(log.StandardLogger())
	rootCMD := server.NewRootCMD()
	if err := rootCMD.Execute(); err != nil {
		log.Errorf("%v\n", err)
		os.Exit(1)
	}
}
