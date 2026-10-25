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
	Name             string
	Total            int
	Success          int
	Failed           int
	Failures         []FailureDetail
	StartTime        time.Time
	EndTime          time.Time
	ExecuteStartTime time.Time
	ExecuteEndTime   time.Time
}

type Checker struct {
	name    string
	query   func(ctx context.Context, startTime, endTime time.Time) ([]string, error)
	account func(transactionId string) error
	logger  logx.Logger
	sem     chan struct{}
}

func NewChecker(name string, query func(ctx context.Context, startTime, endTime time.Time) ([]string, error), account func(transactionId string) error) *Checker {
	return &Checker{
		name:    name,
		query:   query,
		account: account,
		logger:  logx.WithContext(context.Background()),
		sem:     make(chan struct{}, maxConcurrent),
	}
}

func (c *Checker) Check(ctx context.Context, startTime, endTime time.Time) *CheckResult {
	result := &CheckResult{
		Name:      c.name,
		StartTime: startTime,
		EndTime:   endTime,
		Failures:  make([]FailureDetail, 0),
	}

	ids, err := c.query(ctx, startTime, endTime)
	if err != nil {
		c.logger.Errorf("[%s] find pending transfers failed: %v", c.name, err)
		result.Total = 0
		result.Failed = 0
		result.Failures = append(result.Failures, FailureDetail{
			TransactionId: "query_error",
			Error:         fmt.Sprintf("query database failed: %v", err),
		})
		return result
	}

	result.Total = len(ids)
	c.logger.Infof("[%s] found %d pending transfers to check", c.name, result.Total)

	if result.Total == 0 {
		return result
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, id := range ids {
		wg.Add(1)
		c.sem <- struct{}{}
		go func(transactionId string) {
			defer wg.Done()
			defer func() { <-c.sem }()

			maxRetries := 3
			var lastErr error
			for i := 0; i < maxRetries; i++ {
				err := c.account(transactionId)
				if err != nil {
					lastErr = err
					c.logger.Errorf("[%s] call async account failed (retry %d/%d): transaction_id=%s, err=%v",
						c.name, i+1, maxRetries, transactionId, err)
					time.Sleep(time.Duration(i+1) * time.Second)
					continue
				}

				mu.Lock()
				result.Success++
				mu.Unlock()
				c.logger.Infof("[%s] async account success: transaction_id=%s", c.name, transactionId)
				return
			}

			mu.Lock()
			result.Failed++
			result.Failures = append(result.Failures, FailureDetail{
				TransactionId: transactionId,
				Error:         fmt.Sprintf("failed after %d retries: %v", maxRetries, lastErr),
			})
			mu.Unlock()
			c.logger.Errorf("[%s] async account failed after %d retries: transaction_id=%s, err=%v",
				c.name, maxRetries, transactionId, lastErr)
		}(id)
	}

	wg.Wait()
	return result
}

func (r *CheckResult) String() string {
	queryDuration := r.EndTime.Sub(r.StartTime)
	execDuration := r.ExecuteEndTime.Sub(r.ExecuteStartTime)
	result := fmt.Sprintf("=== Check Result [%s] ===\n", r.Name)
	result += fmt.Sprintf("Batch Start Time: %s\n", r.ExecuteStartTime.Format("2006-01-02 15:04:05"))
	result += fmt.Sprintf("Batch End Time:   %s\n", r.ExecuteEndTime.Format("2006-01-02 15:04:05"))
	result += fmt.Sprintf("Batch Duration:   %v\n", execDuration)
	result += fmt.Sprintf("Query Time Range: %s ~ %s\n", r.StartTime.Format("2006-01-02 15:04:05"), r.EndTime.Format("2006-01-02 15:04:05"))
	result += fmt.Sprintf("Query Duration:   %v\n", queryDuration)
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
