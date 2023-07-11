package server

import (
	"os"

	"github.com/bamboovir/cs425/lib/mp3/config"
	"github.com/bamboovir/cs425/lib/mp3/server"
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

func RootCMDMain(serverID string, configPath string) (err error) {
	logger.Infof("serverID: [%s]", serverID)
	serverConfig, err := config.ConfigParser(configPath)
	if err != nil {
		return err
	}

	server, err := server.New(serverID, serverConfig)
	if err != nil {
		return err
	}

	err = server.Start()
	if err != nil {
		return err
	}
	return nil
}

func NewRootCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mp3",
		Short: "mp3",
		Long:  "Distributed transactions::support transactions that read and write to distributed objects while ensuring full ACI(D) properties.",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			nodeID := args[0]
			configPath := args[1]
			err := RootCMDMain(nodeID, configPath)
			ExitWrapper(err)
		},
	}

	return cmd
}
