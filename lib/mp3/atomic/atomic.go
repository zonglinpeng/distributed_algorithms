package atomic

import (
	sync "github.com/sasha-s/go-deadlock"
)

type AtomicInt64 struct {
	data int64
	lock *sync.RWMutex
}

func NewAtomicInt64(data int64) *AtomicInt64 {
	return &AtomicInt64{
		data: data,
		lock: &sync.RWMutex{},
	}
}

func (a *AtomicInt64) Add() {
	defer a.lock.Unlock()
	a.lock.Lock()
	a.data += 1
}

func (a *AtomicInt64) Get() int64 {
	defer a.lock.RUnlock()
	a.lock.RLock()
	return a.data
}

type AtomicBool struct {
	data bool
	lock *sync.RWMutex
}

func NewAtomicBool(data bool) *AtomicBool {
	return &AtomicBool{
		data: data,
		lock: &sync.RWMutex{},
	}
}

func (a *AtomicBool) Get() bool {
	defer a.lock.RUnlock()
	a.lock.RLock()
	return a.data
}

func (a *AtomicBool) Set(data bool) {
	defer a.lock.Unlock()
	a.lock.Lock()
	a.data = data
}
