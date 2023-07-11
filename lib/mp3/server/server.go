package server

import (
	"context"
	"net"

	"github.com/bamboovir/cs425/lib/mp3/bank"
	"github.com/bamboovir/cs425/lib/mp3/config"
	"github.com/bamboovir/cs425/lib/mp3/server/session"
	"github.com/bamboovir/cs425/lib/mp3/transaction"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	logger = log.WithField("src", "server")
)

const (
	CONN_TYPE = "tcp"
	CONN_HOST = "0.0.0.0"
)

type Server struct {
	transaction.UnimplementedTransactionServer
	serverID       string
	addr           string
	bank           *bank.Bank
	sessionManager *session.SessionManager
}

func New(serverID string, nodesConfig *config.Config) (*Server, error) {
	selfConfigItem, err := nodesConfig.FindConfigItemByID(serverID)
	if err != nil {
		return nil, err
	}
	addr := net.JoinHostPort(CONN_HOST, selfConfigItem.NodePort)
	server := &Server{
		serverID:       serverID,
		addr:           addr,
		bank:           bank.New(),
		sessionManager: session.NewSessionManager(),
	}
	return server, nil
}

func (t *Server) Begin(ctx context.Context, req *transaction.BeginReq) (*transaction.BeginRes, error) {
	logger.Infof("received begin: %v", req)
	t.sessionManager.CreateSession(session.GenSessionID(req.ClientID, req.TransactionSeq))
	return &transaction.BeginRes{
		IsOk:           true,
		TransactionSeq: req.TransactionSeq,
	}, nil
}

func (t *Server) Deposit(ctx context.Context, req *transaction.DepositReq) (*transaction.DepositRes, error) {
	logger.Infof("received deposit: %v", req)
	currSession, ok := t.sessionManager.GetSession(session.GenSessionID(req.ClientID, req.TransactionSeq))
	if !ok {
		return &transaction.DepositRes{
			IsOk:           false,
			TransactionSeq: req.TransactionSeq,
		}, nil
	}
	currAccount, ok := currSession.GetAccount(req.Account)

	if ok {
		currAccount.Lock()
		currAccount.SetAmount(currAccount.GetAmount() + req.Amount)
	} else {
		bankAccount := t.bank.GetOrCreateAccount(req.Account)
		currAccount := currSession.CreateAccount(
			req.Account,
			maxInt64(bankAccount.GetAmount(), 0),
			bankAccount.GetLock(),
		)

		currAccount.Lock()
		currAccount.SetAmount(currAccount.GetAmount() + req.Amount)
	}

	return &transaction.DepositRes{
		IsOk:           true,
		TransactionSeq: req.TransactionSeq,
	}, nil
}

func (t *Server) Balance(ctx context.Context, req *transaction.BalanceReq) (*transaction.BalanceRes, error) {
	logger.Infof("received balance: %v", req)
	currSession, ok := t.sessionManager.GetSession(session.GenSessionID(req.ClientID, req.TransactionSeq))
	if !ok {
		return &transaction.BalanceRes{
			IsAccountExist: false,
			Amount:         0,
			TransactionSeq: req.TransactionSeq,
		}, nil
	}
	currAccount, ok := currSession.GetAccount(req.Account)
	if ok {
		currAccount.RLock()
		return &transaction.BalanceRes{
			IsAccountExist: true,
			Amount:         currAccount.GetAmount(),
			TransactionSeq: req.TransactionSeq,
		}, nil
	} else {
		bankAccount, ok := t.bank.GetAccount(req.Account)
		if !ok {
			t.sessionManager.Release(session.GenSessionID(req.ClientID, req.TransactionSeq))
			return &transaction.BalanceRes{
				IsAccountExist: false,
				Amount:         0,
				TransactionSeq: req.TransactionSeq,
			}, nil
		} else {
			if bankAccount.GetAmount() == -1 {
				t.sessionManager.Release(session.GenSessionID(req.ClientID, req.TransactionSeq))
				return &transaction.BalanceRes{
					IsAccountExist: false,
					Amount:         0,
					TransactionSeq: req.TransactionSeq,
				}, nil
			}

			currAccount := currSession.CreateAccount(
				req.Account,
				maxInt64(bankAccount.GetAmount(), 0),
				bankAccount.GetLock(),
			)

			currAccount.RLock()
			return &transaction.BalanceRes{
				IsAccountExist: true,
				Amount:         currAccount.GetAmount(),
				TransactionSeq: req.TransactionSeq,
			}, nil
		}
	}
}

func (t *Server) WithDraw(ctx context.Context, req *transaction.WithDrawReq) (*transaction.WithDrawRes, error) {
	logger.Infof("received withdraw: %v", req)
	currSession, ok := t.sessionManager.GetSession(session.GenSessionID(req.ClientID, req.TransactionSeq))
	if !ok {
		return &transaction.WithDrawRes{
			IsAccountExist: false,
			IsOk:           false,
			TransactionSeq: req.TransactionSeq,
		}, nil
	}
	currAccount, ok := currSession.GetAccount(req.Account)

	if ok {
		currAccount.Lock()
		currAccount.SetAmount(currAccount.GetAmount() - req.Amount)
		return &transaction.WithDrawRes{
			IsAccountExist: true,
			IsOk:           true,
			TransactionSeq: req.TransactionSeq,
		}, nil
	} else {
		bankAccount, ok := t.bank.GetAccount(req.Account)
		if !ok {
			t.sessionManager.Release(session.GenSessionID(req.ClientID, req.TransactionSeq))
			return &transaction.WithDrawRes{
				IsAccountExist: false,
				IsOk:           false,
				TransactionSeq: req.TransactionSeq,
			}, nil
		} else {
			if bankAccount.GetAmount() == -1 {
				t.sessionManager.Release(session.GenSessionID(req.ClientID, req.TransactionSeq))
				return &transaction.WithDrawRes{
					IsAccountExist: false,
					IsOk:           false,
					TransactionSeq: req.TransactionSeq,
				}, nil
			}

			currAccount := currSession.CreateAccount(
				req.Account,
				maxInt64(bankAccount.GetAmount(), 0),
				bankAccount.GetLock(),
			)
			currAccount.Lock()
			currAccount.SetAmount(currAccount.GetAmount() - req.Amount)
			return &transaction.WithDrawRes{
				IsAccountExist: true,
				IsOk:           true,
				TransactionSeq: req.TransactionSeq,
			}, nil
		}
	}
}

func (t *Server) TryCommit(ctx context.Context, req *transaction.TryCommitReq) (*transaction.TryCommitRes, error) {
	logger.Infof("received try commit: %v", req)
	defer t.bank.GetLock().Unlock()
	t.bank.GetLock().Lock()
	currSession, ok := t.sessionManager.GetSession(session.GenSessionID(req.ClientID, req.TransactionSeq))
	if !ok {
		return &transaction.TryCommitRes{
			IsOk:           false,
			TransactionSeq: req.TransactionSeq,
		}, nil
	}

	for accountName, accountInfo := range currSession.GetAccounts() {
		account, _ := t.bank.GetAccount(accountName)
		balance := maxInt64(account.GetAmount(), 0) + accountInfo.Diff()
		if balance < 0 {
			return &transaction.TryCommitRes{
				IsOk:           false,
				TransactionSeq: req.TransactionSeq,
			}, nil
		}
	}

	return &transaction.TryCommitRes{
		IsOk:           true,
		TransactionSeq: req.TransactionSeq,
	}, nil
}

func (t *Server) Commit(ctx context.Context, req *transaction.CommitReq) (*transaction.CommitRes, error) {
	logger.Infof("received commit: %v", req)
	defer t.bank.GetLock().Unlock()
	t.bank.GetLock().Lock()
	currSession, ok := t.sessionManager.GetSession(session.GenSessionID(req.ClientID, req.TransactionSeq))
	if !ok {
		return &transaction.CommitRes{
			IsOk:           false,
			TransactionSeq: req.TransactionSeq,
		}, nil
	}
	for accountName, accountInfo := range currSession.GetAccounts() {
		account, _ := t.bank.GetAccount(accountName)
		balance := maxInt64(account.GetAmount(), 0) + accountInfo.Diff()
		logger.Infof("%s COMMIT %s %d", t.serverID, accountName, balance)
		account.SetAmount(balance)
	}

	t.sessionManager.Release(session.GenSessionID(req.ClientID, req.TransactionSeq))
	return &transaction.CommitRes{
		IsOk:           true,
		TransactionSeq: req.TransactionSeq,
	}, nil
}

func (t *Server) Abort(ctx context.Context, req *transaction.AbortReq) (*transaction.AbortRes, error) {
	logger.Infof("received abort: %v", req)
	t.sessionManager.Release(session.GenSessionID(req.ClientID, req.TransactionSeq))
	return &transaction.AbortRes{
		IsOk:           true,
		TransactionSeq: req.TransactionSeq,
	}, nil
}

func (t *Server) Serve() (err error) {
	socket, err := net.Listen(CONN_TYPE, t.addr)
	if err != nil {
		logger.Errorf("server failed to listen: %v", err)
		return err
	}

	s := grpc.NewServer()
	transaction.RegisterTransactionServer(s, t)
	logger.Infof("server listening at %v", socket.Addr())

	err = s.Serve(socket)
	if err != nil {
		logger.Errorf("server failed to serve: %v", err)
		return err
	}
	return nil
}

func (t *Server) Start() (err error) {
	err = t.Serve()
	if err != nil {
		return err
	}
	return nil
}
