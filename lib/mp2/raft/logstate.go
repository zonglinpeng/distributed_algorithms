package raft

import "fmt"

func (r *Raft) LogTerm(_ int, new int) {
	v := fmt.Sprintf("%d", new)
	r.group.LogState("term", v)
}

func (r *Raft) LogState(_ string, new string) {
	v := fmt.Sprintf("\"%s\"", new)
	r.group.LogState("state", v)
}

func (r *Raft) LogLeader(_ string, new string) {
	if new == "" {
		v := "null"
		r.group.LogState("leader", v)
		return
	}
	v := fmt.Sprintf("\"%s\"", new)
	r.group.LogState("leader", v)
}

func (r *Raft) LogLogEntry(index int, entry LogEntry) {
	k := fmt.Sprintf("log[%d]", index)
	v := fmt.Sprintf("[%d, \"%s\"]", entry.Term, entry.Body)
	r.group.LogState(k, v)
}

func (r *Raft) LogCommitIndex(_ int, new int) {
	v := fmt.Sprintf("%d", new)
	r.group.LogState("commitIndex", v)
}
