package raft

import (
	"encoding/json"

	"github.com/pkg/errors"
)

const (
	AppendEntriesResPath = "AppendEntriesRes"
	RequestVoteResPath   = "RequestVoteRes"
	AppendEntriesReqPath = "AppendEntriesReq"
	RequestVoteReqPath   = "RequestVoteReq"
	LogEntryPath         = "LogEntry"
)

type LogEntry struct {
	Path string `json:"path"`
	Body string `json:"body"`
	Term int    `json:"term"`
}

type AppendEntriesReq struct {
	Path         string     `json:"path"`
	NID          string     `json:"nid"`
	Term         int        `json:"term"`
	LeaderID     string     `json:"leader_id"`
	PrevLogIndex int        `json:"prev_log_index"`
	PrevLogTerm  int        `json:"prev_log_term"`
	Entries      []LogEntry `json:"entries"`
	LeaderCommit int        `json:"leader_commit"`
}

type AppendEntriesRes struct {
	Path      string `json:"path"`
	NID       string `json:"nid"`
	Term      int    `json:"term"`
	IsSuccess bool   `json:"is_success"`
}

type RequestVoteReq struct {
	Path         string `json:"path"`
	NID          string `json:"nid"`
	Term         int    `json:"term"`
	CandidateID  string `json:"candidate_id"`
	LastLogIndex int    `json:"last_log_index"`
	LastLogTerm  int    `json:"last_log_term"`
}

type RequestVoteRes struct {
	Path        string `json:"path"`
	NID         string `json:"nid"`
	Term        int    `json:"term"`
	VoteGranted bool   `json:"vote_granted"`
}

func NewLogEntry() *LogEntry {
	return &LogEntry{
		Path: LogEntryPath,
	}
}

func NewAppendEntriesReq() *AppendEntriesReq {
	return &AppendEntriesReq{
		Path: AppendEntriesReqPath,
	}
}

func NewAppendEntriesRes() *AppendEntriesRes {
	return &AppendEntriesRes{
		Path: AppendEntriesResPath,
	}
}

func NewRequestVoteReq() *RequestVoteReq {
	return &RequestVoteReq{
		Path: RequestVoteReqPath,
	}
}

func NewRequestVoteRes() *RequestVoteRes {
	return &RequestVoteRes{
		Path: RequestVoteResPath,
	}
}

func (m *LogEntry) Decode(data []byte) (msg *LogEntry, err error) {
	err = json.Unmarshal(data, m)
	if err != nil {
		return nil, errors.Wrap(err, "decode logEntry failed")
	}
	return m, nil
}

func (m *AppendEntriesReq) Decode(data []byte) (msg *AppendEntriesReq, err error) {
	err = json.Unmarshal(data, m)
	if err != nil {
		return nil, errors.Wrap(err, "decode AppendEntriesReq failed")
	}
	return m, nil
}

func (m *AppendEntriesRes) Decode(data []byte) (msg *AppendEntriesRes, err error) {
	err = json.Unmarshal(data, m)
	if err != nil {
		return nil, errors.Wrap(err, "decode AppendEntriesRes failed")
	}
	return m, nil
}

func (m *RequestVoteReq) Decode(data []byte) (msg *RequestVoteReq, err error) {
	err = json.Unmarshal(data, m)
	if err != nil {
		return nil, errors.Wrap(err, "decode RequestVoteReq failed")
	}
	return m, nil
}

func (m *RequestVoteRes) Decode(data []byte) (msg *RequestVoteRes, err error) {
	err = json.Unmarshal(data, m)
	if err != nil {
		return nil, errors.Wrap(err, "decode RequestVoteRes failed")
	}
	return m, nil
}

func (m *LogEntry) Equal(data *LogEntry) bool {
	return m.Path == data.Path && m.Term == data.Term && m.Body == data.Body
}
