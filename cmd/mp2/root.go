package mp2

import (
	"os"
	"strconv"

	"github.com/zonglinpeng/distributed_algorithms/lib/mp2/raft"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	logger = log.WithField("src", "main")
)

func ExitWrapper(err error) {
	if err != nil {
		logger.Errorf("command err: %v", err)
		os.Exit(1)
	}
}

func ParseRootCMDParams(nidRawStr string, numOfNodesRawStr string) (nid string, numOfNodes int, err error) {
	numOfNodes, err = strconv.Atoi(numOfNodesRawStr)
	if err != nil {
		return "", -1, errors.Wrap(err, "parse num of nodes failed")
	}
	return nidRawStr, numOfNodes, nil
}

func RootCMDMain(nidRawStr string, numOfNodesRawStr string) (err error) {
	nid, numOfNodes, err := ParseRootCMDParams(nidRawStr, numOfNodesRawStr)
	if err != nil {
		return errors.Wrap(err, "parse root cmd params failed")
	}
	logger.Infof("start with nid [%s], num of nodes [%d], pid [%d]", nid, numOfNodes, os.Getpid())
	raft := raft.New(nid, numOfNodes)
	raft.Start()
	select {}
}

func NewRootCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "raft",
		Short: "raft",
		Long:  "raft",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			nidRawStr := args[0]
			numOfNodesRawStr := args[1]
			err := RootCMDMain(nidRawStr, numOfNodesRawStr)
			ExitWrapper(err)
		},
	}

	return cmd
}
