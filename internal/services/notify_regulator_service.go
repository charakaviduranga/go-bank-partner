package services

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/array/banking-api/internal/models"
	"github.com/google/uuid"
)

// NotifyRegulatorService handles polling transfer status and notifying regulator
type NotifyRegulatorService struct {
	northWindService NorthWindServiceInterface
	auditLogger      AuditLoggerInterface
	timeout          time.Duration
	pollInterval     time.Duration
	maxRetries       int
	logger           *slog.Logger
}

// NewNotifyRegulator creates a new transfer monitor
func NewNotifyRegulator(northWindService NorthWindServiceInterface, auditLogger AuditLoggerInterface, logger *slog.Logger) NotifyRegulatorServiceInterface {
	return &NotifyRegulatorService{
		northWindService: northWindService,
		auditLogger:      auditLogger,
		logger:           logger,
		timeout:          60 * time.Second,
		pollInterval:     1 * time.Second,
		maxRetries:       5,
	}
}

// StartMonitoring starts polling transfer status and notifying regulator
func (m *NotifyRegulatorService) StartMonitoring(transferID uuid.UUID, fromAccount, toAccount, amount string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
		defer cancel()

		finalStatus := m.PollTransferStatus(ctx, transferID)
		m.logger.Info("Transfer Final Status", "transferID", transferID)
		// Audit log the final status
		if m.auditLogger != nil {
			m.auditLogger.LogEvent(ctx, "Transfer final status detected", map[string]interface{}{
				"transfer_id": transferID,
				"status":      finalStatus,
			})
		}

		// Notify regulator
		m.NotifyRegulator(ctx, transferID, finalStatus == models.TransferStatusCompleted, map[string]interface{}{
			"amount":       amount,
			"from_account": fromAccount,
			"to_account":   toAccount,
			"status":       finalStatus,
		})
	}()
}

// pollTransferStatus polls the transfer status until completed, failed, or timeout
func (m *NotifyRegulatorService) PollTransferStatus(ctx context.Context, transferID uuid.UUID) string {
	for {
		select {
		case <-ctx.Done():
			return models.TransferStatusPending
		default:
			status, err := m.northWindService.GetTransferStatus(ctx, transferID.String())
			if err != nil {
				time.Sleep(m.pollInterval)
				continue
			}
			if strings.ToLower(status.Status) == models.TransferStatusCompleted || strings.ToLower(status.Status) == models.TransferStatusFailed {
				return status.Status
			}
			time.Sleep(m.pollInterval)
		}
	}
}

// notifyRegulator sends webhook with retries inside the timeout window
func (m *NotifyRegulatorService) NotifyRegulator(ctx context.Context, transferID uuid.UUID, success bool, details map[string]interface{}) {
	start := time.Now()
	payload := map[string]interface{}{
		"transfer_id": transferID.String(),
		"success":     success,
		"details":     details,
	}

	for attempt := 1; attempt <= m.maxRetries; attempt++ {
		remaining := m.timeout - time.Since(start)
		if remaining <= 0 {
			if m.auditLogger != nil {
				m.auditLogger.LogEvent(ctx, "Regulator notify failed: 60s timeout", map[string]interface{}{
					"transfer_id": transferID,
				})
			}
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), remaining)
		err := m.northWindService.Notify(ctx, payload)
		cancel()

		if err == nil {
			if m.auditLogger != nil {
				m.auditLogger.LogEvent(ctx, "Regulator notified successfully", map[string]interface{}{
					"transfer_id": transferID,
					"attempt":     attempt,
				})
			}
			return
		}

		if m.auditLogger != nil {
			m.auditLogger.LogEvent(ctx, "Regulator notify failed", map[string]interface{}{
				"transfer_id": transferID,
				"attempt":     attempt,
				"error":       err.Error(),
			})
		}

		sleep := time.Second * time.Duration(attempt)
		if sleep > remaining {
			sleep = remaining
		}
		time.Sleep(sleep)
	}
}
