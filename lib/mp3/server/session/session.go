package session

import (
	"strconv"
	"time"

	"github.com/bamboovir/cs425/lib/mp3/atomic"
	"github.com/bamboovir/cs425/lib/mp3/rwlock"
	sync "github.com/sasha-s/go-deadlock"

	log "github.com/sirupsen/logrus"
)

var (
	logger = log.WithField("src", "session")
)

type Session struct {
	accounts map[string]*AccountInfo
	lock     *sync.RWMutex
}

type AccountInfo struct {
	amount       int64
	amountLock   *sync.Mutex
	originAmount int64
	lock         *rwlock.RWMutex
	setRLock     *atomic.AtomicBool
	setWLock     *atomic.AtomicBool
	released     chan struct{}
}

type SessionManager struct {
	sessions map[string]*Session
	lock     *sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: map[string]*Session{},
		lock:     &sync.RWMutex{},
	}
}

func NewSession() *Session {
	return &Session{
		accounts: map[string]*AccountInfo{},
		lock:     &sync.RWMutex{},
	}
}

func NewAccountInfo(amount int64, lock *rwlock.RWMutex) *AccountInfo {
	released := make(chan struct{}, 1)
	a := &AccountInfo{
		amount:       amount,
		amountLock:   &sync.Mutex{},
		lock:         lock,
		originAmount: amount,
		setRLock:     atomic.NewAtomicBool(false),
		setWLock:     atomic.NewAtomicBool(false),
	}

	go func() {
		<-released
		ticker := time.NewTicker(time.Second * 1)

		if a.Unlock() {
			logger.Infof("release write lock")
			return
		}

		if a.RUnlock() {
			logger.Infof("release read lock")
			return
		}

		for range ticker.C {
			if a.Unlock() {
				logger.Infof("release write lock")
				return
			}
			if a.RUnlock() {
				logger.Infof("release read lock")
				return
			}
		}
	}()

	a.released = released
	return a
}

func (s *SessionManager) CreateSession(clientID string) {
	defer s.lock.Unlock()
	s.lock.Lock()
	s.sessions[clientID] = NewSession()
}

func (s *SessionManager) GetSession(clientID string) (*Session, bool) {
	defer s.lock.RUnlock()
	s.lock.RLock()
	session, ok := s.sessions[clientID]
	return session, ok
}

func (s *SessionManager) Release(clientID string) {
	defer s.lock.Unlock()
	s.lock.Lock()
	session, ok := s.sessions[clientID]
	if ok {
		session.Release()
		delete(s.sessions, clientID)
	}
}

func (s *Session) GetAccount(accountName string) (*AccountInfo, bool) {
	defer s.lock.RUnlock()
	s.lock.RLock()

	account, ok := s.accounts[accountName]
	return account, ok
}

func (s *Session) CreateAccount(accountName string, amount int64, lock *rwlock.RWMutex) *AccountInfo {
	defer s.lock.Unlock()
	s.lock.Lock()
	account := NewAccountInfo(amount, lock)
	s.accounts[accountName] = account
	return account
}

func (s *Session) GetAccounts() map[string]*AccountInfo {
	return s.accounts
}

func (s *Session) Release() {
	defer s.lock.RUnlock()
	s.lock.RLock()
	for _, accountInfo := range s.accounts {
		accountInfo.Release()
	}
}

func (a *AccountInfo) Release() {
	a.released <- struct{}{}
}

func (a *AccountInfo) GetAmount() int64 {
	defer a.amountLock.Unlock()
	a.amountLock.Lock()
	return a.amount
}

func (a *AccountInfo) SetAmount(amount int64) {
	defer a.amountLock.Unlock()
	a.amountLock.Lock()
	a.amount = amount
}

func (a *AccountInfo) Diff() int64 {
	defer a.amountLock.Unlock()
	a.amountLock.Lock()
	return a.amount - a.originAmount
}

func (a *AccountInfo) RLock() {
	if a.setRLock.Get() {
		return
	}
	if a.setWLock.Get() {
		return
	}
	a.lock.RLock()
	a.setRLock.Set(true)
}

func (a *AccountInfo) Lock() {
	if a.setWLock.Get() {
		return
	}
	if a.setRLock.Get() {
		a.lock.Upgrade()
		a.setWLock.Set(true)
		return
	}
	a.lock.Lock()
	a.setWLock.Set(true)
}

func (a *AccountInfo) RUnlock() bool {
	if !a.setRLock.Get() {
		return false
	}

	a.lock.RUnlock()
	return true
}

func (a *AccountInfo) Unlock() bool {
	if !a.setWLock.Get() {
		return false
	}

	a.lock.Unlock()
	return true
}

func GenSessionID(clientID string, transactionSeq int64) string {
	return clientID + strconv.FormatInt(transactionSeq, 10)
}
