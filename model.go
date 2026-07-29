package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var ErrNotFound = sql.ErrNoRows

type TC2cPendingTransfer struct {
	TransactionId string    `db:"transaction_id"`
	BuyerUid      int64     `db:"buyer_uid"`
	SellerUid     int64     `db:"seller_uid"`
	BuyerUserId   string    `db:"buyer_user_id"`
	SellerUserId  string    `db:"seller_user_id"`
	Amount        int64     `db:"amount"`
	State         int64     `db:"state"`
	Desc          string    `db:"desc"`
	CreateTime    time.Time `db:"create_time"`
	UpdateTime    time.Time `db:"update_time"`
}

type C2cPendingTransferModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewC2cPendingTransferModel(conn sqlx.SqlConn) *C2cPendingTransferModel {
	return &C2cPendingTransferModel{
		conn:  conn,
		table: "`t_c2c_pending_transfer`",
	}
}

func (m *C2cPendingTransferModel) FindPendingByTimeRange(ctx context.Context, startTime, endTime time.Time) ([]*TC2cPendingTransfer, error) {
	query := fmt.Sprintf(`
		SELECT transaction_id, buyer_uid, seller_uid, buyer_user_id, seller_user_id, 
		       amount, state, desc, create_time, update_time
		FROM %s 
		WHERE state = 1 
		  AND create_time >= ? 
		  AND create_time < ?
	`, m.table)

	var resp []*TC2cPendingTransfer
	err := m.conn.QueryRowsCtx(ctx, &resp, query, startTime, endTime)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
