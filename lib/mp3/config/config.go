package config

import (
	"bufio"
	"os"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	logger = log.WithField("src", "config")
)

type Config struct {
	ConfigItems []ConfigItem
}

type ConfigItem struct {
	NodeID   string
	NodeHost string
	NodePort string
}

func (c *Config) FindConfigItemByID(nid string) (*ConfigItem, error) {
	for _, item := range c.ConfigItems {
		if item.NodeID == nid {
			return &item, nil
		}
	}

	return nil, errors.Errorf("couldn't find config item by nid [%s]", nid)
}

func ConfigParser(path string) (config *Config, err error) {
	configItems, err := ConfigItemsParser(path)
	if err != nil {
		return nil, err
	}
	return &Config{
		ConfigItems: configItems,
	}, nil
}

func ConfigItemsParser(path string) (configItem []ConfigItem, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	configItem = make([]ConfigItem, 0)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		fields := strings.Fields(line)
		if len(fields) != 3 {
			logger.Errorf("invalid input format, skip")
			continue
		}

		nodeIDStr := fields[0]
		nodeHostStr := fields[1]
		nodePortStr := fields[2]

		logger.Infof("append config [%s][%s][%s]", nodeIDStr, nodeHostStr, nodePortStr)

		configItem = append(configItem, ConfigItem{
			NodeID:   nodeIDStr,
			NodeHost: nodeHostStr,
			NodePort: nodePortStr,
		})
	}

	err = scanner.Err()

	if err != nil {
		return nil, err
	}

	return configItem, nil
}
