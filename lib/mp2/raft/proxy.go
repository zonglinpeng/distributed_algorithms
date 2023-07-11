package raft

import (
	"fmt"
	"strings"
)

type IntProxy struct {
	value int
	f     func(old int, new int)
}

type StrProxy struct {
	value string
	f     func(old string, new string)
}

type LogEntriesProxy struct {
	entries []LogEntry
	f       func(index int, entry LogEntry)
}

func NewIntProxy(v int, f func(old int, new int)) *IntProxy {
	return &IntProxy{
		value: v,
		f:     f,
	}
}

func NewStrProxy(v string, f func(old string, new string)) *StrProxy {
	return &StrProxy{
		value: v,
		f:     f,
	}
}

func NewLogEntriesProxy(f func(index int, entry LogEntry)) *LogEntriesProxy {
	return &LogEntriesProxy{
		entries: make([]LogEntry, 0),
		f:       f,
	}
}

func (p *IntProxy) GetInt() int {
	return p.value
}

func (p *StrProxy) GetStr() string {
	return p.value
}

func (p *IntProxy) SetInt(v int) {
	if p.value == v {
		return
	} else {
		p.f(p.value, v)
		p.value = v
	}
}

func (p *StrProxy) SetStr(v string) {
	if p.value == v {
		return
	} else {
		p.f(p.value, v)
		p.value = v
	}
}

func (p *LogEntriesProxy) Append(entry LogEntry) {
	p.entries = append(p.entries, entry)
	p.f(len(p.entries), entry)
}

func (p *LogEntriesProxy) Shrink(start int, end int) {
	p.entries = p.entries[start:end]
}

func (p *LogEntriesProxy) Slice(start int, end int) []LogEntry {
	return p.entries[start:end]
}

func (p *LogEntriesProxy) Len() int {
	return len(p.entries)
}

func (p *LogEntriesProxy) Get(index int) LogEntry {
	return p.entries[index]
}

func (p *LogEntriesProxy) Set(index int, entry LogEntry) {
	p.entries[index] = entry
}

func (p *LogEntriesProxy) Snapshot() string {
	buf := &strings.Builder{}

	for _, entry := range p.entries {
		curr := fmt.Sprintf("[%d@%s], ", entry.Term, entry.Body)
		buf.WriteString(curr)
	}
	
	return buf.String()
}
