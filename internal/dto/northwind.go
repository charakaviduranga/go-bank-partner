package dto

import (
	"time"

	"github.com/array/banking-api/internal/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ---------- Validation ----------

type ValidationResponse[T any] struct {
	Validation Validation `json:"validation"`
	Data       *T         `json:"data,omitempty"`
}

type Validation struct {
	Valid          bool              `json:"valid"`
	Issues         []ValidationIssue `json:"issues,omitempty"`
	Metadata       map[string]any    `json:"metadata,omitempty"`
	ValidationTime time.Time         `json:"validation_time"`
}

type ValidationIssue struct {
	Field    string `json:"field"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

// ---------- Account ----------

type AccountValidationRequest struct {
	AccountHolderName string `json:"account_holder_name"`
	AccountNumber     string `json:"account_number"`
	RoutingNumber     string `json:"routing_number"`
}

type Account struct {
	ID               string          `json:"account_id"`
	AccountHolder    string          `json:"account_holder_name"`
	Status           string          `json:"account_status"`
	Type             string          `json:"account_type"`
	AvailableBalance decimal.Decimal `json:"available_balance"`
}

type AccountValidationResult struct {
	Response         *ValidationResponse[Account]
	AccountExists    bool
	AccountValid     bool
	AvailableBalance decimal.Decimal
}

// ---------- Transfer ----------

type TransferAccount struct {
	AccountHolderName string `json:"account_holder_name"`
	AccountNumber     string `json:"account_number"`
	RoutingNumber     string `json:"routing_number"`
	InstitutionName   string `json:"institution_name,omitempty"`
}

type TransferValidationRequest struct {
	Amount             float64         `json:"amount"`
	Currency           string          `json:"currency"`
	Direction          string          `json:"direction"`
	TransferType       string          `json:"transfer_type"`
	Description        string          `json:"description"`
	ReferenceNumber    string          `json:"reference_number"`
	ScheduledDate      string          `json:"scheduled_date"`
	SourceAccount      TransferAccount `json:"source_account"`
	DestinationAccount TransferAccount `json:"destination_account"`
}

type TransferInitiationRequest = TransferValidationRequest

type TransferStatusEvent struct {
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
}

type TransferBetweenAccountsInput struct {
	FromAccountID  uuid.UUID
	ToAccountID    uuid.UUID
	Amount         decimal.Decimal
	Description    string
	IdempotencyKey string
	Currency       string
	TransferType   string
	Direction      string
	UserID         uuid.UUID
	Institution    string
}

type TransferValidationInput struct {
	ReferenceNumber string
	SourceUser      *models.User
	DestinationUser *models.User
	FromAccount     *models.Account
	ToAccount       *models.Account
	Amount          decimal.Decimal
	Description     string
}

type TransferStatus struct {
	TransferID             string                `json:"transfer_id"`
	Status                 string                `json:"status"`
	ReferenceNumber        string                `json:"reference_number"`
	Amount                 decimal.Decimal       `json:"amount"`
	Currency               string                `json:"currency"`
	Direction              string                `json:"direction"`
	TransferType           string                `json:"transfer_type"`
	Description            string                `json:"description"`
	InitiatedDate          time.Time             `json:"initiated_date"`
	ExpectedCompletionDate time.Time             `json:"expected_completion_date"`
	Fee                    decimal.Decimal       `json:"fee"`
	ExchangeRate           decimal.Decimal       `json:"exchange_rate"`
	RetryCount             int                   `json:"retry_count"`
	SourceAccount          TransferAccount       `json:"source_account"`
	DestinationAccount     TransferAccount       `json:"destination_account"`
	StatusHistory          []TransferStatusEvent `json:"status_history"`
}

// ---------- Errors ----------

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details"`
	RequestID string         `json:"request_id"`
	Timestamp string         `json:"timestamp"`
}

type BalanceResponse struct {
	AccountNumber    string          `json:"account_number"`
	AvailableBalance decimal.Decimal `json:"available_balance"`
	Currency         string          `json:"currency"`
	CurrentBalance   decimal.Decimal `json:"current_balance"`
	LastUpdated      string          `json:"last_updatedI"`
}
