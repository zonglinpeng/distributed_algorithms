package group

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	LogEntryPath = "LogEntry"
)

var (
	logger = log.WithField("src", "group")
)

type Group struct {
	nid        string
	numOfNodes int
	routers    map[string]func(ctx *RecvContext)
}

type RecvContext struct {
	NID  string
	Path string
	Data []byte
}

func (r RecvContext) String() string {
	return fmt.Sprintf("nid %s; path %s; data %s", r.NID, r.Path, string(r.Data))
}

type PayloadPath struct {
	Path string `json:"path"`
}

func RecvContextFromFrameWorkLine(line string) (ctx *RecvContext, err error) {
	// RECEIVE ${nid} ${payload:json}
	// tmp[0] tmp[1] tmp[2]
	tmp := strings.Fields(line)
	if len(tmp) != 3 && len(tmp) != 2 {
		logger.Errorf("invalid input format, skip")
		return nil, errors.New("invalid input format, skip")
	}

	if len(tmp) == 3 {
		if tmp[0] == "RECEIVE" {
			nid := tmp[1]
			payloadStr := tmp[2]
			payloadBytes := []byte(payloadStr)
			payload := &PayloadPath{}

			err = json.Unmarshal(payloadBytes, payload)
			if err != nil {
				logger.Errorf("invalid input format, skip")
				return nil, errors.Wrap(err, "invalid input format, skip")
			}

			return &RecvContext{
				NID:  nid,
				Path: payload.Path,
				Data: payloadBytes,
			}, nil
		} else {
			logger.Errorf("invalid input format, skip")
			return nil, errors.New("invalid input format, skip")
		}
	}

	if len(tmp) == 2 {
		if tmp[0] == "LOG" {
			payloadStr := tmp[1]
			payloadBytes := []byte(payloadStr)

			// payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadStr)
			// if err != nil {
			// 	logger.Errorf("invalid input format, skip")
			// 	return nil, errors.Wrap(err, "invalid input format, skip")
			// }

			return &RecvContext{
				Path: LogEntryPath,
				Data: payloadBytes,
			}, nil
		} else {
			logger.Errorf("invalid input format, skip")
			return nil, errors.New("invalid input format, skip")
		}
	}

	logger.Errorf("invalid input format, skip")
	return nil, errors.New("invalid input format, skip")
}

func New(nid string, numOfNodes int) *Group {
	return &Group{
		nid:        nid,
		numOfNodes: numOfNodes,
		routers:    map[string]func(ctx *RecvContext){},
	}
}

func (g *Group) LogState(variableName string, value string) (err error) {
	payload := fmt.Sprintf("STATE %s=%s\n", variableName, value)
	_, err = os.Stdout.WriteString(payload)
	if err != nil {
		return errors.Wrap(err, "log framework state failed")
	}
	return nil
}

func (g *Group) LogCommitted(value string, index int) (err error) {
	// encoded := base64.RawURLEncoding.EncodeToString([]byte(value))
	encoded := value
	payload := fmt.Sprintf("COMMITTED %s %d\n", encoded, index)
	_, err = os.Stdout.WriteString(payload)
	if err != nil {
		return errors.Wrap(err, "log framework committed failed")
	}
	return nil
}

func (g *Group) Unicast(dstID string, v interface{}) (err error) {
	payload, err := EncodeFrameworkMsg(dstID, v)
	if err != nil {
		return errors.Wrap(err, "unicast framework msg failed")
	}
	_, err = os.Stdout.WriteString(payload)
	if err != nil {
		return errors.Wrap(err, "unicast framework msg failed")
	}
	return nil
}

func (g *Group) BMulticast(v interface{}) (err error) {
	for i := 0; i < g.numOfNodes; i++ {
		dstID := strconv.Itoa(i)
		err := g.Unicast(dstID, v)
		if err != nil {
			return errors.Wrap(err, "b multicast failed")
		}
	}
	return nil
}

func (g *Group) BMulticastWithoutSelf(v interface{}) (err error) {
	for i := 0; i < g.numOfNodes; i++ {
		dstID := strconv.Itoa(i)
		if dstID == g.nid {
			continue
		}
		err := g.Unicast(dstID, v)
		if err != nil {
			return errors.Wrap(err, "b multicast failed")
		}
	}
	return nil
}

func EncodeFrameworkMsg(dstID string, v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", errors.Wrap(err, "encode framework msg failed")
	}
	return fmt.Sprintf("SEND %s %s\n", dstID, string(data)), nil
}

func (g *Group) NID() string {
	return g.nid
}

func (g *Group) NumOfNodes() int {
	return g.numOfNodes
}

func (g *Group) Route(path string, f func(ctx *RecvContext)) {
	g.routers[path] = f
}

func (g *Group) HandleRoute(ctx *RecvContext) {
	f, ok := g.routers[ctx.Path]
	if !ok {
		errmsg := fmt.Sprintf("ctx %s don't match any router", ctx)
		logger.Errorf("%s", errmsg)
		return
	}
	f(ctx)
}

func (g *Group) Start() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		// logger.Infof("RECV %s", line)
		ctx, err := RecvContextFromFrameWorkLine(line)
		if err != nil {
			logger.Errorf("invalid input format, skip %+v", err)
			continue
		}
		g.HandleRoute(ctx)
	}

	err := scanner.Err()

	if err != nil {
		logger.Errorf("read err: %v", err)
	} else {
		logger.Info("reach EOF")
	}
}
