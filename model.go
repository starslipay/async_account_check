package main

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type PendingTransferModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewC2cPendingTransferModel(conn sqlx.SqlConn) *PendingTransferModel {
	return &PendingTransferModel{conn: conn, table: "`t_c2c_pending_transfer`"}
}

func NewC2bPendingTransferModel(conn sqlx.SqlConn) *PendingTransferModel {
	return &PendingTransferModel{conn: conn, table: "`t_c2b_pending_transfer`"}
}

func (m *PendingTransferModel) FindPendingIdsByTimeRange(ctx context.Context, startTime, endTime time.Time) ([]string, error) {
	query := fmt.Sprintf("SELECT transaction_id FROM %s WHERE state = 1 AND create_time >= ? AND create_time < ?", m.table)

	var ids []string
	err := m.conn.QueryRowsCtx(ctx, &ids, query, startTime, endTime)
	if err != nil {
		return nil, err
	}
	return ids, nil
}
