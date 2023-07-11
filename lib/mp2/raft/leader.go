package raft

import (
	"strconv"
	"time"
)

func (r *Raft) Leader() string {
	r.currentLeader.SetStr(r.nid)
	r.resetNextIndex()
	r.bMulticastAppendEntriesReq()

	ticker := time.NewTicker(HeartbeatIntervals * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.logSnapshot("ticker.C")
			r.logger.Infof("leader send heartbeat to followers")
			r.bMulticastAppendEntriesReq()
		case msg := <-r.appendEntriesReqChan:
			r.logSnapshot("appendEntriesReq")
			if msg.Term <= r.currentTerm.GetInt() {
				continue
			}

			r.currentLeader.SetStr("")
			r.currentTerm.SetInt(msg.Term)
			r.currentLeader.SetStr(msg.NID)

			r.votedFor = ""

			return FollowerState

		case msg := <-r.appendEntriesResChan:
			r.logSnapshot("appendEntriesRes")

			if msg.Term > r.currentTerm.GetInt() {
				r.updateTerm(msg.Term)
				return FollowerState
			}

			nid, _ := strconv.Atoi(msg.NID)

			if msg.IsSuccess {
				r.matchIndex[nid] = r.nextIndex[nid]
				r.nextIndex[nid] += 1
				matchIndex := r.matchIndex[nid]
				r.tryIncCommitIndex(matchIndex, msg.Term)
			} else {
				r.nextIndex[nid] -= 1
				rmsg := r.BuildAppendEntriesReq(nid)
				r.group.Unicast(msg.NID, rmsg)
			}
			continue
		case msg := <-r.requestVoteReqChan:
			r.logSnapshot("requestVoteReq")

			if msg.Term <= r.currentTerm.GetInt() ||
				!r.moreOrEqualUpToDateLog(msg.LastLogIndex, msg.LastLogTerm) {
				rmsg := NewRequestVoteRes()
				rmsg.Term = r.currentTerm.GetInt()
				rmsg.VoteGranted = false
				r.group.Unicast(msg.CandidateID, rmsg)
				continue
			}

			r.currentLeader.SetStr("")
			r.currentTerm.SetInt(msg.Term)
			r.votedFor = msg.CandidateID

			rmsg := NewRequestVoteRes()
			rmsg.Term = r.currentTerm.GetInt()
			rmsg.VoteGranted = true

			r.group.Unicast(msg.CandidateID, rmsg)
			return FollowerState
		case msg := <-r.requestVoteResChan:
			r.logSnapshot("requestVoteRes")

			if msg.Term > r.currentTerm.GetInt() {
				r.updateTerm(msg.Term)
				return FollowerState
			}
			continue
		case msg := <-r.logEntryChan:
			r.logSnapshot("logEntry")
			logEntry := NewLogEntry()
			logEntry.Body = msg.Body
			logEntry.Term = r.currentTerm.GetInt()
			r.logs.Append(*logEntry)
			continue
		}
	}
}

func (r *Raft) BuildAppendEntriesReq(node int) *AppendEntriesReq {
	msg := NewAppendEntriesReq()
	msg.LeaderID = r.nid
	msg.Term = r.currentTerm.GetInt()
	msg.LeaderCommit = r.commitIndex.GetInt()
	nextIndex := r.nextIndex[node]
	prevIndex := nextIndex - 1

	if prevIndex == -1 {
		msg.PrevLogIndex = -1
		msg.PrevLogTerm = -1
	} else {
		msg.PrevLogIndex = prevIndex
		msg.PrevLogTerm = r.logs.Get(prevIndex).Term
	}

	lower := maxInt(0, nextIndex)
	upper := minInt(r.logs.Len(), nextIndex+1)
	msg.Entries = r.logs.Slice(lower, upper)
	return msg
}

func (r *Raft) bMulticastAppendEntriesReq() {
	for i := 0; i < r.numOfNodes; i++ {
		dstID := strconv.Itoa(i)
		if dstID == r.nid {
			continue
		}

		rmsg := r.BuildAppendEntriesReq(i)
		r.group.Unicast(dstID, rmsg)
	}
}

func (r *Raft) updateTerm(term int) bool {
	if term > r.currentTerm.GetInt() {
		r.currentLeader.SetStr("")
		r.currentTerm.SetInt(term)
		r.votedFor = ""
		return true
	}
	return false
}

func (r *Raft) resetNextIndex() {
	for i := range r.nextIndex {
		r.nextIndex[i] = maxInt(0, r.logs.Len()-1)
	}
}

func (r *Raft) tryIncCommitIndex(matchIndex int, term int) {
	if (matchIndex + 1) < r.commitIndex.GetInt() {
		return
	}
	
	count := 1
	for _, v := range r.matchIndex {
		if v >= matchIndex {
			count += 1
		}
	}

	if count >= r.majorityNum &&
		r.logs.Get(matchIndex).Term == term {
		r.commitIndex.SetInt(matchIndex + 1)
		r.apply(r.logs.Get(matchIndex).Body, matchIndex + 1)
	}
}

func (r *Raft) apply(op string, index int) {
	r.group.LogCommitted(op, index)
	r.lastApplied = index
}
