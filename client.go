package main

import (
	"context"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/zeromicro/go-zero/zrpc"
)

type AccountMgrClient struct {
	cli zrpc.Client
}

func NewAccountMgrClient(conf zrpc.RpcClientConf) *AccountMgrClient {
	conn := zrpc.MustNewClient(conf)
	return &AccountMgrClient{
		cli: conn,
	}
}

func (c *AccountMgrClient) C2CAsyncAccount(transactionId string) error {
	client := account_mgr_pb.NewAccountMgrClient(c.cli.Conn())
	_, err := client.C2CAsyncAccount(context.Background(), &account_mgr_pb.C2CAsyncAccountReq{
		TransactionId: transactionId,
	})
	return err
}

func (c *AccountMgrClient) C2BAsyncAccount(transactionId string) error {
	client := account_mgr_pb.NewAccountMgrClient(c.cli.Conn())
	_, err := client.C2BAsyncAccount(context.Background(), &account_mgr_pb.C2BAsyncAccountReq{
		TransactionId: transactionId,
	})
	return err
}
