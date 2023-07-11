package rwlock

import (
	sync "github.com/sasha-s/go-deadlock"
)

type RWMutex struct {
	rwlock *sync.RWMutex
	lock   *sync.Mutex
}

func New() *RWMutex {
	return &RWMutex{
		rwlock: &sync.RWMutex{},
		lock:   &sync.Mutex{},
	}
}

func (l *RWMutex) Lock() {
	defer l.lock.Unlock()
	l.lock.Lock()

	l.rwlock.Lock()
}

func (l *RWMutex) Unlock() {
	l.rwlock.Unlock()
}

func (l *RWMutex) RLock() {
	defer l.lock.Unlock()
	l.lock.Lock()

	l.rwlock.RLock()
}

func (l *RWMutex) RUnlock() {
	l.rwlock.RUnlock()
}

func (l *RWMutex) Upgrade() {
	defer l.lock.Unlock()
	l.lock.Lock()

	l.rwlock.RUnlock()
	l.rwlock.Lock()
}

func (l *RWMutex) Downgrade() {
	defer l.lock.Unlock()
	l.lock.Lock()

	l.rwlock.Unlock()
	l.rwlock.RLock()
}
