package raft

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/zonglinpeng/distributed_algorithms/lib/mp2/group"
	log "github.com/sirupsen/logrus"
)

const (
	FollowerState  = "FOLLOWER"
	CandidateState = "CANDIDATE"
	LeaderState    = "LEADER"
	// ElectionTimeoutMax = 300
	// ElectionTimeoutMin = 150
	// HeartbeatIntervals = 30
	ElectionTimeoutMax = 3000
	ElectionTimeoutMin = 1500
	HeartbeatIntervals = 300
	ChannelBufferSize  = 100000
)

type Raft struct {
	rand                 *rand.Rand
	logger               *log.Entry
	nid                  string
	numOfNodes           int
	majorityNum          int
	group                *group.Group
	requestVoteResChan   chan *RequestVoteRes
	requestVoteReqChan   chan *RequestVoteReq
	appendEntriesReqChan chan *AppendEntriesReq
	appendEntriesResChan chan *AppendEntriesRes
	logEntryChan         chan *LogEntry
	currentTerm          *IntProxy
	currentLeader        *StrProxy
	votedFor             string
	logs                 *LogEntriesProxy
	commitIndex          *IntProxy
	state                *StrProxy
	lastApplied          int
	nextIndex            []int
	matchIndex           []int
}

func New(nid string, numOfNodes int) *Raft {
	group := group.New(nid, numOfNodes)
	r := &Raft{}
	r.logger = log.WithField("src", "raft").WithField("nid", nid)
	r.nid = nid
	r.numOfNodes = numOfNodes
	r.majorityNum = numOfNodes/2 + 1
	r.group = group
	r.requestVoteResChan = make(chan *RequestVoteRes, ChannelBufferSize)
	r.requestVoteReqChan = make(chan *RequestVoteReq, ChannelBufferSize)
	r.appendEntriesReqChan = make(chan *AppendEntriesReq, ChannelBufferSize)
	r.appendEntriesResChan = make(chan *AppendEntriesRes, ChannelBufferSize)
	r.logEntryChan = make(chan *LogEntry, ChannelBufferSize)
	r.currentTerm = NewIntProxy(0, r.LogTerm)
	r.currentLeader = NewStrProxy("", r.LogLeader)
	r.votedFor = ""
	r.logs = NewLogEntriesProxy(r.LogLogEntry)
	r.commitIndex = NewIntProxy(0, r.LogCommitIndex)
	r.state = NewStrProxy(FollowerState, r.LogState)
	r.lastApplied = 0
	r.nextIndex = make([]int, numOfNodes)
	r.matchIndex = make([]int, numOfNodes)
	r.rand = r.initRand()
	return r
}

func (r *Raft) AppendEntriesReqRoute(ctx *group.RecvContext) {
	// r.logger.Infof("recv AppendEntriesReq %s", ctx)
	msg := NewAppendEntriesReq()
	_, err := msg.Decode(ctx.Data)
	if err != nil {
		r.logger.Errorf("AppendEntriesReqRoute:: %+v", err)
		return
	}
	msg.NID = ctx.NID
	r.appendEntriesReqChan <- msg
}

func (r *Raft) AppendEntriesResRoute(ctx *group.RecvContext) {
	// r.logger.Infof("recv AppendEntriesRes %s", ctx)
	msg := NewAppendEntriesRes()
	_, err := msg.Decode(ctx.Data)
	if err != nil {
		r.logger.Errorf("AppendEntriesResRoute:: %+v", err)
		return
	}
	msg.NID = ctx.NID
	r.appendEntriesResChan <- msg
}

func (r *Raft) RequestVoteReqRoute(ctx *group.RecvContext) {
	// r.logger.Infof("recv RequestVoteReq %s", ctx)
	msg := NewRequestVoteReq()
	_, err := msg.Decode(ctx.Data)
	if err != nil {
		r.logger.Errorf("RequestVoteReqRoute:: %+v", err)
		return
	}
	msg.NID = ctx.NID
	r.requestVoteReqChan <- msg
}

func (r *Raft) RequestVoteResRoute(ctx *group.RecvContext) {
	// r.logger.Infof("recv RequestVoteRes %s", ctx)
	msg := NewRequestVoteRes()
	_, err := msg.Decode(ctx.Data)
	if err != nil {
		r.logger.Errorf("RequestVoteResRoute:: %+v", err)
		return
	}
	msg.NID = ctx.NID
	r.requestVoteResChan <- msg
}

func (r *Raft) LogEntryRoute(ctx *group.RecvContext) {
	// r.logger.Infof("recv RequestVoteRes %s", ctx)
	msg := NewLogEntry()
	msg.Path = ctx.Path
	msg.Body = string(ctx.Data)
	msg.Term = r.currentTerm.GetInt()
	r.logEntryChan <- msg
}

func (r *Raft) Start() {
	r.group.Route(AppendEntriesResPath, r.AppendEntriesResRoute)
	r.group.Route(AppendEntriesReqPath, r.AppendEntriesReqRoute)
	r.group.Route(RequestVoteResPath, r.RequestVoteResRoute)
	r.group.Route(RequestVoteReqPath, r.RequestVoteReqRoute)
	r.group.Route(LogEntryPath, r.LogEntryRoute)

	go r.group.Start()

	for {
		switch r.state.GetStr() {
		case FollowerState:
			r.state.SetStr(r.Follower())
			r.logger.Infof("state:: %s -> %s", FollowerState, r.state.GetStr())
			continue
		case CandidateState:
			r.state.SetStr(r.Candidate())
			r.logger.Infof("state:: %s -> %s", CandidateState, r.state.GetStr())
			continue
		case LeaderState:
			r.state.SetStr(r.Leader())
			r.logger.Infof("state:: %s -> %s", LeaderState, r.state.GetStr())
			continue
		default:
			panic(fmt.Sprintf("raft state is invalid: %s", r.state.GetStr()))
		}
	}
}

func (r *Raft) logSnapshot(event string) {
	r.logger.Infof(
		">>> event: [%s] | self::[%s] | state::[%s] | term::[%d] | leader::[%s] | commit_index::[%d] | vote_for::[%s] | logs size: [%d] | logs: [%s] | next index: [%s] | match index: [%s]",
		event,
		r.nid,
		r.state.GetStr(),
		r.currentTerm.GetInt(),
		r.currentLeader.GetStr(),
		r.commitIndex.GetInt(),
		r.votedFor,
		r.logs.Len(),
		r.logs.Snapshot(),
		r.SnapshotNextIndex(),
		r.SnapshotMatchIndex(),
	)
}

func (r *Raft) SnapshotNextIndex() string {
	buf := &strings.Builder{}

	for i, entry := range r.nextIndex {
		curr := fmt.Sprintf("[%d@%d], ", i, entry)
		buf.WriteString(curr)
	}

	return buf.String()
}

func (r *Raft) SnapshotMatchIndex() string {
	buf := &strings.Builder{}

	for i, entry := range r.matchIndex {
		curr := fmt.Sprintf("[%d@%d], ", i, entry)
		buf.WriteString(curr)
	}

	return buf.String()
}
