package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/zonglinpeng/distributed_algorithms/lib/mp3/atomic"
	"github.com/zonglinpeng/distributed_algorithms/lib/mp3/client"
	"github.com/zonglinpeng/distributed_algorithms/lib/mp3/config"
	"github.com/zonglinpeng/distributed_algorithms/lib/mp3/transaction"
	"github.com/zonglinpeng/distributed_algorithms/lib/mp3/transaction/command"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	logger = log.WithField("src", "main")
)

func ExitWrapper(err error) {
	if err != nil {
		logger.Errorf("command err: %v", err)
		os.Exit(1)
	}
}

func HandleEventFromReader(reader io.Reader, clientManager *client.ClientManager) (err error) {
	transactionSeq := atomic.NewAtomicInt64(0)
	ctx := context.Background()
	clientID := clientManager.ClientID
	scanner := bufio.NewScanner(reader)
	isInTransaction := atomic.NewAtomicBool(false)
	forceSeq := make(chan struct{}, 1)
	forceSeq <- struct{}{}

	for scanner.Scan() {
		line := scanner.Text()
		event, err := command.ParseLine(line)
		if err != nil {
			logger.Errorf("encode input msg failed with err :%v, skip", err)
			continue
		}

		clients := clientManager.GetAllClients()
		switch event.(type) {
		case *command.Abort:
			if !isInTransaction.Get() {
				logger.Error("not in transaction, skip abort event")
				continue
			}

			logger.Infof("Abort Event")
			for serverID, client := range clients {
				rsp, err := client.Abort(ctx, &transaction.AbortReq{
					ClientID:       clientID,
					TransactionSeq: transactionSeq.Get(),
				})
				if err != nil {
					logger.Errorf("process abort event failed by server: %v, continue", serverID, err)
					continue
				}
				logger.Infof("abort event process ok by server [%s] with rsp: %v", serverID, rsp)
			}
			fmt.Printf("ABORTED\n")
			isInTransaction.Set(false)
			transactionSeq.Add()
			select {
			case forceSeq <- struct{}{}:
			default:
			}
			continue
		default:
			<-forceSeq
			go func() {
				defer func() {
					select {
					case forceSeq <- struct{}{}:
					default:
					}
				}()
				switch e := event.(type) {
				case *command.Begin:
					if isInTransaction.Get() {
						logger.Error("already in transaction, skip begin event")
						return
					}
					isInTransaction.Set(true)
					transactionSeq.Add()
					logger.Infof("Begin Event")
					for serverID, client := range clients {
						rsp, err := client.Begin(ctx, &transaction.BeginReq{
							ClientID:       clientID,
							TransactionSeq: transactionSeq.Get(),
						})
						if err != nil {
							logger.Errorf("process begin event failed by server [%s]: %v, continue", serverID, err)
							return
						}
						logger.Infof("begin event process ok by server [%s] with rsp: %v", serverID, rsp)
					}
					fmt.Printf("OK\n")
					return
				case *command.Deposit:
					if !isInTransaction.Get() {
						logger.Error("not in transaction, skip deposit event")
						return
					}
					logger.Infof("Deposit Event: [%s][%s][%d]", e.Server, e.Account, e.Amount)
					client, ok := clientManager.GetClientByID(e.Server)
					if !ok {
						logger.Errorf("server [%s] not exist", e.Server)
						return
					}
					rsp, err := client.Deposit(ctx, &transaction.DepositReq{
						ClientID:       clientID,
						Server:         e.Server,
						Account:        e.Account,
						Amount:         e.Amount,
						TransactionSeq: transactionSeq.Get(),
					})
					if err != nil {
						logger.Errorf("process deposit event failed: %v, continue", err)
						return
					}
					logger.Infof("deposit event process ok with rsp: %v", rsp)
					if rsp.TransactionSeq != transactionSeq.Get() {
						logger.Infof("transaction seq not match, skip")
						return
					}
					fmt.Printf("OK\n")
					return
				case *command.Balance:
					if !isInTransaction.Get() {
						logger.Error("not in transaction, skip balance event")
						return
					}
					logger.Infof("Balance Event: [%s][%s]", e.Server, e.Account)
					client, ok := clientManager.GetClientByID(e.Server)
					if !ok {
						logger.Errorf("server [%s] not exist", e.Server)
						return
					}
					rsp, err := client.Balance(ctx, &transaction.BalanceReq{
						ClientID:       clientID,
						Server:         e.Server,
						Account:        e.Account,
						TransactionSeq: transactionSeq.Get(),
					})
					if err != nil {
						logger.Errorf("process balance event failed: %v, continue", err)
						return
					}
					logger.Infof("balance event process ok with rsp: %v", rsp)
					if rsp.TransactionSeq != transactionSeq.Get() {
						logger.Infof("transaction seq not match, skip")
						return
					}
					if !rsp.IsAccountExist {
						fmt.Printf("NOT FOUND, ABORTED\n")
						isInTransaction.Set(false)
						return
					}
					fmt.Printf("%s.%s = %d\n", e.Server, e.Account, rsp.Amount)
					return
				case *command.WithDraw:
					if !isInTransaction.Get() {
						logger.Error("not in transaction, skip withdraw event")
						return
					}
					logger.Infof("WithDraw Event: [%s][%s][%d]", e.Server, e.Account, e.Amount)
					client, ok := clientManager.GetClientByID(e.Server)
					if !ok {
						logger.Errorf("server [%s] not exist", e.Server)
						return
					}
					rsp, err := client.WithDraw(ctx, &transaction.WithDrawReq{
						ClientID:       clientID,
						Server:         e.Server,
						Account:        e.Account,
						Amount:         e.Amount,
						TransactionSeq: transactionSeq.Get(),
					})
					if err != nil {
						logger.Errorf("process withdraw event failed: %v, continue", err)
						return
					}
					logger.Infof("withdraw event process ok with rsp: %v", rsp)
					if rsp.TransactionSeq != transactionSeq.Get() {
						logger.Infof("transaction seq not match, skip")
						return
					}
					if !rsp.IsAccountExist {
						fmt.Printf("NOT FOUND, ABORTED\n")
						isInTransaction.Set(false)
						return
					}
					fmt.Printf("OK\n")
					return
				case *command.Commit:
					if !isInTransaction.Get() {
						logger.Error("not in transaction, skip commit event")
						return
					}
					logger.Infof("Commit Event")
					shouldAbort := false
					for serverID, client := range clients {
						rsp, err := client.TryCommit(ctx, &transaction.TryCommitReq{
							ClientID:       clientID,
							TransactionSeq: transactionSeq.Get(),
						})
						if err != nil {
							logger.Errorf("process try commit event failed by server [%s]: %v, continue", serverID, err)
							return
						}
						logger.Infof("try commit event process ok by server [%s] with rsp: %v", serverID, rsp)
						if !rsp.IsOk {
							shouldAbort = true
							break
						}
					}

					if shouldAbort {
						for serverID, client := range clients {
							rsp, err := client.Abort(ctx, &transaction.AbortReq{
								ClientID:       clientID,
								TransactionSeq: transactionSeq.Get(),
							})
							if err != nil {
								logger.Errorf("process abort event failed by server: %v, continue", serverID, err)
								continue
							}
							logger.Infof("abort event process ok by server [%s] with rsp: %v", serverID, rsp)
						}
						fmt.Printf("ABORTED\n")
						isInTransaction.Set(false)
						return
					}

					for serverID, client := range clients {
						rsp, err := client.Commit(ctx, &transaction.CommitReq{
							ClientID:       clientID,
							TransactionSeq: transactionSeq.Get(),
						})
						if err != nil {
							logger.Errorf("process commit event failed by server [%s]: %v, continue", serverID, err)
							return
						}
						logger.Infof("commit event process ok by server [%s] with rsp: %v", serverID, rsp)
					}
					fmt.Printf("COMMIT OK\n")
					isInTransaction.Set(false)
					return
				default:
					logger.Errorf("invalid event: [%v]", e)
					return
				}
			}()
		}
	}

	err = scanner.Err()

	if err != nil {
		logger.Errorf("reader read err: %v", err)
		return err
	}
	logger.Info("reader reach eof")
	return nil
}

func RootCMDMain(clientID string, configPath string) (err error) {
	logger.Infof("clientID: [%s]", clientID)
	serverConfigs, err := config.ConfigParser(configPath)
	if err != nil {
		return err
	}
	client, err := client.New(clientID, serverConfigs)
	if err != nil {
		return err
	}

	err = HandleEventFromReader(os.Stdin, client)
	if err != nil {
		return err
	}
	return nil
}

func NewRootCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mp3-c",
		Short: "mp3-c",
		Long:  "Client distributed transactions::support transactions that read and write to distributed objects while ensuring full ACI(D) properties.",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			clientID := args[0]
			configPath := args[1]
			err := RootCMDMain(clientID, configPath)
			ExitWrapper(err)
		},
	}

	return cmd
}
