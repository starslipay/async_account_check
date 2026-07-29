package main

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	AccountMgrRpc zrpc.RpcClientConf
	DB            sqlx.SqlConf
	CheckInterval int64 `json:"CheckInterval"`
	TimeStart     int64 `json:"TimeStart"`
	TimeEnd       int64 `json:"TimeEnd"`
}
