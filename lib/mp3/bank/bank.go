package bank

import (
	"github.com/bamboovir/cs425/lib/mp3/rwlock"
	sync "github.com/sasha-s/go-deadlock"

	log "github.com/sirupsen/logrus"
)

var (
	logger = log.WithField("src", "transaction")
)

type AccountInfo struct {
	amount int64
	lock   *rwlock.RWMutex
}

func (a *AccountInfo) GetAmount() int64 {
	return a.amount
}

func (a *AccountInfo) GetLock() *rwlock.RWMutex {
	return a.lock
}

func (a *AccountInfo) SetAmount(amount int64) {
	a.amount = amount
}

type Bank struct {
	balances map[string]*AccountInfo
	lock     *sync.RWMutex
}

func New() *Bank {
	return &Bank{
		balances: map[string]*AccountInfo{},
		lock:     &sync.RWMutex{},
	}
}

func (b *Bank) GetLock() *sync.RWMutex {
	return b.lock
}

func NewAccountInfo() *AccountInfo {
	return &AccountInfo{
		amount: -1,
		lock:   rwlock.New(),
	}
}

func (b *Bank) GetOrCreateAccount(accountName string) *AccountInfo {
	account, ok := b.balances[accountName]
	if ok {
		return account
	}
	account = NewAccountInfo()
	b.balances[accountName] = account
	return account
}

func (b *Bank) GetAccount(accountName string) (*AccountInfo, bool) {
	account, ok := b.balances[accountName]
	return account, ok
}
