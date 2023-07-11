package command

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// BEGIN: Open a new transaction, and reply with “OK”.

// DEPOSIT server.account amount: Deposit some amount into an account.
// Amount will be a positive integer.
// (You can assume that the value of any account will never exceed 1,000,000,000.)
// The account balance should increase by the given amount.
// If the account was previously unreferenced,
// it should be created with an initial balance of amount.
// The client should reply with OK

// BALANCE server.account: The client should display the current balance in the given account

// WITHDRAW server.account amount: Withdraw some amount from an account.
// The account balance should decrease by the withdrawn amount.
// The client should reply with OK if the operation is successful.
// If the account does not exist (i.e, has never received any deposits),
// the client should print NOT FOUND, ABORTED and abort the transaction.

// COMMIT: Commit the transaction, making its results visible to other transactions.
// The client should reply either with COMMIT OK or ABORTED,
// in the case that the transaction had to be aborted during the commit process.

const (
	BeginEvent    = "BEGIN"
	DepositEvent  = "DEPOSIT"
	BalanceEvent  = "BALANCE"
	WithDrawEvent = "WITHDRAW"
	CommitEvent   = "COMMIT"
	AbortEvent    = "ABORT"
)

type Begin struct{}
type Deposit struct {
	Server  string
	Account string
	Amount  int64
}
type Balance struct {
	Server  string
	Account string
}
type WithDraw struct {
	Server  string
	Account string
	Amount  int64
}
type Commit struct{}
type Abort struct{}

func ParseLine(line string) (msg interface{}, err error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		errMsg := "invalid event format, empty fields"
		return nil, errors.New(errMsg)
	}

	eventType := fields[0]

	switch eventType {
	case BeginEvent:
		if len(fields) != 1 {
			errMsg := "invalid begin event format"
			return nil, errors.New(errMsg)
		}
		return &Begin{}, nil
	case DepositEvent:
		if len(fields) != 3 {
			errMsg := "invalid deposit event format"
			return nil, errors.New(errMsg)
		}

		amount, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, err
		}
		firstFields := strings.Split(fields[1], ".")
		if len(firstFields) != 2 {
			errMsg := "invalid deposit event format"
			return nil, errors.New(errMsg)
		}
		server := firstFields[0]
		account := firstFields[1]
		return &Deposit{Server: server, Account: account, Amount: int64(amount)}, nil
	case BalanceEvent:
		if len(fields) != 2 {
			errMsg := "invalid balance event format"
			return nil, errors.New(errMsg)
		}

		firstFields := strings.Split(fields[1], ".")
		if len(firstFields) != 2 {
			errMsg := "invalid balance event format"
			return nil, errors.New(errMsg)
		}
		server := firstFields[0]
		account := firstFields[1]

		return &Balance{Server: server, Account: account}, nil
	case WithDrawEvent:
		if len(fields) != 3 {
			errMsg := "invalid withdraw event format"
			return nil, errors.New(errMsg)
		}

		amount, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, err
		}
		firstFields := strings.Split(fields[1], ".")
		if len(firstFields) != 2 {
			errMsg := "invalid withdraw event format"
			return nil, errors.New(errMsg)
		}
		server := firstFields[0]
		account := firstFields[1]
		return &WithDraw{Server: server, Account: account, Amount: int64(amount)}, nil
	case CommitEvent:
		if len(fields) != 1 {
			errMsg := "invalid commit event format"
			return nil, errors.New(errMsg)
		}
		return &Commit{}, nil
	case AbortEvent:
		if len(fields) != 1 {
			errMsg := "invalid abort event format"
			return nil, errors.New(errMsg)
		}
		return &Abort{}, nil
	default:
		errMsg := "unrecognized event type"
		return nil, errors.New(errMsg)
	}
}
