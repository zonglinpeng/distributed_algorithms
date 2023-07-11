package raft

import "time"

func (r *Raft) Candidate() string {
	// start new election
	r.currentLeader.SetStr("")
	r.currentTerm.SetInt(r.currentTerm.GetInt() + 1)
	r.votedFor = r.nid
	voteCount := 1

	msg := NewRequestVoteReq()
	msg.Term = r.currentTerm.GetInt()
	msg.CandidateID = r.nid
	msg.LastLogIndex = r.lastLogIndex()
	msg.LastLogTerm = r.lastLogTerm()

	r.group.BMulticastWithoutSelf(msg)

	timeout := r.randomElectionTimeout()
	r.logger.Infof("candidate election set timeout:: %s", timeout)
	timer := time.NewTimer(timeout)

	for {
		select {
		case <-timer.C:
			r.logSnapshot("timer.C")
			// times out, start new election
			r.logger.Infof("candidate election timeout after:: %s", timeout)
			return CandidateState
		case msg := <-r.appendEntriesReqChan:
			r.logSnapshot("appendEntriesReq")
			if msg.Term < r.currentTerm.GetInt() {
				continue
			}

			r.currentLeader.SetStr("")
			r.currentTerm.SetInt(msg.Term)
			r.currentLeader.SetStr(msg.LeaderID)

			r.votedFor = ""

			return FollowerState
		case msg := <-r.appendEntriesResChan:
			r.logSnapshot("appendEntriesRes")
			if msg.Term <= r.currentTerm.GetInt() {
				continue
			}

			r.currentLeader.SetStr("")
			r.currentTerm.SetInt(msg.Term)
			r.votedFor = ""
			return FollowerState
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

			if msg.Term < r.currentTerm.GetInt() {
				continue
			}
			if msg.Term > r.currentTerm.GetInt() {
				r.updateTerm(msg.Term)
				return FollowerState
			}

			if msg.VoteGranted {
				r.logger.Infof("candidate get new vote from %s", msg.NID)
				voteCount += 1
				if voteCount >= r.majorityNum {
					return LeaderState
				}
			}
			continue
		case <-r.logEntryChan:
			r.logSnapshot("logEntry")
			r.logger.Infof("ignore log entry")
			continue
		}
	}
}

func (r *Raft) lastLogIndex() int {
	if r.logs.Len() == 0 {
		return -1
	} else {
		return r.logs.Len() - 1
	}
}

func (r *Raft) lastLogTerm() int {
	if r.logs.Len() == 0 {
		return -1
	} else {
		entry := r.logs.Get(r.logs.Len() - 1)
		return entry.Term
	}
}

func (r *Raft) moreOrEqualUpToDateLog(lastLogIndex int, lastLogTerm int) bool {
	selfLastLogIndex := r.lastLogIndex()
	selfLastLogTerm := r.lastLogTerm()
	if selfLastLogTerm != lastLogTerm {
		return selfLastLogTerm > lastLogTerm
	} else {
		return selfLastLogIndex >= lastLogIndex
	}
}
