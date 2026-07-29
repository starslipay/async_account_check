package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

const maxConcurrent = 10

type FailureDetail struct {
	TransactionId string
	Error         string
}

type CheckResult struct {
	Total     int
	Success   int
	Failed    int
	Failures  []FailureDetail
	StartTime time.Time
	EndTime   time.Time
}

type C2CChecker struct {
	accountMgrClient *AccountMgrClient
	model            *C2cPendingTransferModel
	logger           logx.Logger
	sem              chan struct{}
}

func NewC2CChecker(accountMgrClient *AccountMgrClient, model *C2cPendingTransferModel) *C2CChecker {
	return &C2CChecker{
		accountMgrClient: accountMgrClient,
		model:            model,
		logger:           logx.WithContext(context.Background()),
		sem:              make(chan struct{}, maxConcurrent),
	}
}

func (c *C2CChecker) Check(ctx context.Context, startTime, endTime time.Time) *CheckResult {
	result := &CheckResult{
		StartTime: startTime,
		EndTime:   endTime,
		Failures:  make([]FailureDetail, 0),
	}

	transfers, err := c.model.FindPendingByTimeRange(ctx, startTime, endTime)
	if err != nil {
		c.logger.Errorf("find pending transfers failed: %v", err)
		result.Total = 0
		result.Failed = 0
		result.Failures = append(result.Failures, FailureDetail{
			TransactionId: "query_error",
			Error:         fmt.Sprintf("query database failed: %v", err),
		})
		return result
	}

	result.Total = len(transfers)
	c.logger.Infof("found %d pending transfers to check", result.Total)

	if result.Total == 0 {
		return result
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, transfer := range transfers {
		wg.Add(1)
		c.sem <- struct{}{}
		go func(t *TC2cPendingTransfer) {
			defer wg.Done()
			defer func() { <-c.sem }()

			maxRetries := 3
			var lastErr error
			for i := 0; i < maxRetries; i++ {
				_, err := c.accountMgrClient.C2CAsyncAccount(t.TransactionId)
				if err != nil {
					lastErr = err
					c.logger.Errorf("call C2CAsyncAccount failed (retry %d/%d): transaction_id=%s, err=%v",
						i+1, maxRetries, t.TransactionId, err)
					time.Sleep(time.Duration(i+1) * time.Second)
					continue
				}

				mu.Lock()
				result.Success++
				mu.Unlock()
				c.logger.Infof("c2c async account success: transaction_id=%s", t.TransactionId)
				return
			}

			mu.Lock()
			result.Failed++
			result.Failures = append(result.Failures, FailureDetail{
				TransactionId: t.TransactionId,
				Error:         fmt.Sprintf("failed after %d retries: %v", maxRetries, lastErr),
			})
			mu.Unlock()
			c.logger.Errorf("c2c async account failed after %d retries: transaction_id=%s, err=%v",
				maxRetries, t.TransactionId, lastErr)
		}(transfer)
	}

	wg.Wait()
	return result
}

func (r *CheckResult) String() string {
	duration := r.EndTime.Sub(r.StartTime)
	result := fmt.Sprintf("=== Check Result ===\n")
	result += fmt.Sprintf("Start Time: %s\n", r.StartTime.Format(time.RFC3339))
	result += fmt.Sprintf("End Time: %s\n", r.EndTime.Format(time.RFC3339))
	result += fmt.Sprintf("Duration: %v\n", duration)
	result += fmt.Sprintf("Total: %d\n", r.Total)
	result += fmt.Sprintf("Success: %d\n", r.Success)
	result += fmt.Sprintf("Failed: %d\n", r.Failed)

	if r.Failed > 0 && len(r.Failures) > 0 {
		result += fmt.Sprintf("\nFailure Details:\n")
		for i, f := range r.Failures {
			result += fmt.Sprintf("  %d. transaction_id=%s, error=%s\n", i+1, f.TransactionId, f.Error)
		}
	}

	result += fmt.Sprintf("====================\n")
	return result
}
