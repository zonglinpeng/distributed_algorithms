package raft

import "time"

func (r *Raft) Follower() string {
	timeout := r.randomElectionTimeout()
	r.logger.Infof("follower election set timeout:: %s", timeout)
	timer := time.NewTimer(timeout)
	for {
		select {
		case <-timer.C:
			r.logSnapshot("timer.C")
			r.logger.Infof("follower election timeout after:: %s", timeout)
			return CandidateState
		case msg := <-r.appendEntriesReqChan:
			r.logSnapshot("appendEntriesReq")

			if msg.Term < r.currentTerm.GetInt() {
				continue
			}

			if msg.Term > r.currentTerm.GetInt() {
				r.votedFor = ""
				r.currentLeader.SetStr("")
				r.currentTerm.SetInt(msg.Term)
				r.currentLeader.SetStr(msg.LeaderID)
			}

			if !r.isLogEntryMatch(msg.PrevLogIndex, msg.PrevLogTerm) {
				rmsg := NewAppendEntriesRes()
				rmsg.Term = r.currentTerm.GetInt()
				rmsg.IsSuccess = false
				r.group.Unicast(msg.NID, rmsg)
				return FollowerState
			}

			r.currentLeader.SetStr(msg.LeaderID)

			r.alignLogEntries(msg.PrevLogIndex, msg.Entries)

			if msg.LeaderCommit > r.commitIndex.GetInt() && r.logs.Len() == msg.LeaderCommit {
				r.commitIndex.SetInt(msg.LeaderCommit)
			}

			rmsg := NewAppendEntriesRes()
			rmsg.Term = r.currentTerm.GetInt()
			rmsg.IsSuccess = true

			if len(msg.Entries) != 0 {
				r.group.Unicast(msg.NID, rmsg)
			}

			return FollowerState
		case msg := <-r.appendEntriesResChan:
			r.logSnapshot("appendEntriesRes")

			if msg.Term > r.currentTerm.GetInt() {
				r.currentLeader.SetStr("")
				r.currentTerm.SetInt(msg.Term)
				r.votedFor = ""
				return FollowerState
			}
			continue
		case msg := <-r.requestVoteReqChan:
			r.logSnapshot("requestVoteReq")

			if msg.Term < r.currentTerm.GetInt() ||
				!r.moreOrEqualUpToDateLog(msg.LastLogIndex, msg.LastLogTerm) {
				rmsg := NewRequestVoteRes()
				rmsg.Term = r.currentTerm.GetInt()
				rmsg.VoteGranted = false
				r.group.Unicast(msg.CandidateID, rmsg)
				continue
			}

			if msg.Term > r.currentTerm.GetInt() {
				r.updateTerm(msg.Term)
			}

			rmsg := NewRequestVoteRes()
			rmsg.Term = r.currentTerm.GetInt()
			if r.votedFor == "" || r.votedFor == msg.CandidateID {
				r.votedFor = msg.CandidateID
				rmsg.VoteGranted = true
				r.group.Unicast(msg.CandidateID, rmsg)
				return FollowerState
			} else {
				rmsg.VoteGranted = false
				r.group.Unicast(msg.CandidateID, rmsg)
			}

			continue
		case msg := <-r.requestVoteResChan:
			r.logSnapshot("requestVoteRes")

			if msg.Term > r.currentTerm.GetInt() {
				r.updateTerm(msg.Term)
				return FollowerState
			}
			continue
		case <-r.logEntryChan:
			r.logSnapshot("logEntry")
			r.logger.Infof("ignore log entry")
			continue
		}
	}
}

func (r *Raft) isLogEntryMatch(prevLogIndex int, prevLogTerm int) bool {
	if prevLogIndex == -1 && prevLogTerm == -1 {
		return true
	}

	if prevLogIndex < 0 || prevLogIndex >= r.logs.Len() {
		return false
	}
	logEntry := r.logs.Get(prevLogIndex)
	return logEntry.Term == prevLogTerm
}

func (r *Raft) alignLogEntries(prevLogIndex int, entries []LogEntry) {
	startIndex := prevLogIndex + 1
	for _, entry := range entries {
		if startIndex >= r.logs.Len() {
			r.logs.Append(entry)
			startIndex += 1
			continue
		}

		currEntry := r.logs.Get(startIndex)

		if entry.Term != currEntry.Term {
			r.logs.Shrink(0, startIndex)
			r.logs.Append(entry)
			startIndex += 1
			continue
		} else {
			r.logs.Set(startIndex, currEntry)
			startIndex += 1
			continue
		}
	}
}
