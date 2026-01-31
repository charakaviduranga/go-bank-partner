package services

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/array/banking-api/internal/config"
	"github.com/array/banking-api/internal/dto"
	"github.com/array/banking-api/internal/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

type NorthWindServiceTestSuite struct {
	suite.Suite
	logger  *slog.Logger
	service *NorthWindService
	config  *config.NorthWindConfig
}

func (suite *NorthWindServiceTestSuite) SetupTest() {
	suite.logger = slog.New(slog.NewTextHandler(os.Stderr, nil))

	suite.config = &config.NorthWindConfig{
		BaseURL:    "https://api.northwind.local",
		APIKey:     "test-api-key",
		WebhookURL: "https://webhook.local/notify",
	}

	suite.service = &NorthWindService{
		cfg:    suite.config,
		client: &http.Client{Timeout: 10 * time.Second},
		logger: suite.logger,
	}
}

// ============================================================================
// AuthAccount Tests
// ============================================================================

func (suite *NorthWindServiceTestSuite) TestAuthAccount_Success() {
	requestDto := dto.AccountValidationRequest{
		AccountHolderName: "John Doe",
		AccountNumber:     "1234567890",
		RoutingNumber:     "021000021",
	}

	validationResponse := dto.ValidationResponse[dto.Account]{
		Validation: dto.Validation{Valid: true},
		Data: &dto.Account{
			ID:               "acc-123",
			AvailableBalance: decimal.NewFromFloat(5000.00),
		},
	}

	respBody, _ := json.Marshal(validationResponse)

	originalClient := suite.service.client
	suite.service.client = &http.Client{
		Transport: &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(respBody)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			},
		},
	}
	defer func() { suite.service.client = originalClient }()

	result, err := suite.service.ValidateAccount(context.Background(), requestDto)

	suite.NoError(err)
	suite.NotNil(result)
	suite.True(result.AccountValid)
	suite.True(result.AccountExists)
	suite.Equal("acc-123", result.Response.Data.ID)
	suite.True(result.AvailableBalance.Equal(decimal.NewFromFloat(5000.00)))
}

func (suite *NorthWindServiceTestSuite) TestAuthAccount_BadRequest() {
	requestDto := dto.AccountValidationRequest{
		AccountHolderName: "John Doe",
		AccountNumber:     "invalid",
		RoutingNumber:     "021000021",
	}

	errResponse := dto.ErrorResponse{
		Error: dto.ErrorDetail{
			Code:      "VALIDATION_ERROR",
			Message:   "Invalid account number",
			RequestID: "req-123",
		},
	}

	respBody, _ := json.Marshal(errResponse)

	originalClient := suite.service.client
	suite.service.client = &http.Client{
		Transport: &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(bytes.NewReader(respBody)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			},
		},
	}
	defer func() { suite.service.client = originalClient }()

	result, err := suite.service.ValidateAccount(context.Background(), requestDto)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "Invalid account number")
}

func (suite *NorthWindServiceTestSuite) TestAuthAccount_UnauthorizedAPIKey() {
	requestDto := dto.AccountValidationRequest{
		AccountHolderName: "John Doe",
		AccountNumber:     "1234567890",
		RoutingNumber:     "021000021",
	}

	errResponse := dto.ErrorResponse{
		Error: dto.ErrorDetail{
			Code:      "AUTH_ERROR",
			Message:   "Unauthorized: Invalid API key",
			RequestID: "req-456",
		},
	}

	respBody, _ := json.Marshal(errResponse)

	originalClient := suite.service.client
	suite.service.client = &http.Client{
		Transport: &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusUnauthorized,
					Body:       io.NopCloser(bytes.NewReader(respBody)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			},
		},
	}
	defer func() { suite.service.client = originalClient }()

	result, err := suite.service.ValidateAccount(context.Background(), requestDto)

	suite.Error(err)
	suite.Nil(result)
}

func (suite *NorthWindServiceTestSuite) TestAuthAccount_ServerError() {
	requestDto := dto.AccountValidationRequest{
		AccountHolderName: "John Doe",
		AccountNumber:     "1234567890",
		RoutingNumber:     "021000021",
	}

	errResponse := dto.ErrorResponse{
		Error: dto.ErrorDetail{
			Code:      "SERVER_ERROR",
			Message:   "Internal server error",
			RequestID: "req-789",
		},
	}

	respBody, _ := json.Marshal(errResponse)

	originalClient := suite.service.client
	suite.service.client = &http.Client{
		Transport: &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewReader(respBody)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			},
		},
	}
	defer func() { suite.service.client = originalClient }()

	result, err := suite.service.ValidateAccount(context.Background(), requestDto)

	suite.Error(err)
	suite.Nil(result)
}

func (suite *NorthWindServiceTestSuite) TestAuthAccount_UnexpectedStatusCode() {
	requestDto := dto.AccountValidationRequest{
		AccountHolderName: "John Doe",
		AccountNumber:     "1234567890",
		RoutingNumber:     "021000021",
	}

	originalClient := suite.service.client
	suite.service.client = &http.Client{
		Transport: &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusTeapot,
					Body:       io.NopCloser(bytes.NewReader([]byte("I'm a teapot"))),
					Header:     http.Header{"Content-Type": []string{"text/plain"}},
				}, nil
			},
		},
	}
	defer func() { suite.service.client = originalClient }()

	result, err := suite.service.ValidateAccount(context.Background(), requestDto)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "northwind error (418)")
}

// ============================================================================
// ValidateTransfer Tests
// ============================================================================

func (suite *NorthWindServiceTestSuite) TestValidateTransfer_Success() {
	sourceUser := &models.User{
		ID:        uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
		FirstName: "John",
		LastName:  "Doe",
	}

	destinationUser := &models.User{
		ID:        uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d480"),
		FirstName: "Jane",
		LastName:  "Smith",
	}

	fromAccount := &models.Account{
		ID:            uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d481"),
		AccountNumber: "1000000001",
		RoutingNumber: "021000021",
	}

	toAccount := &models.Account{
		ID:            uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d482"),
		AccountNumber: "1000000002",
		RoutingNumber: "021000021",
	}

	validationResponse := dto.ValidationResponse[dto.TransferStatus]{
		Validation: dto.Validation{Valid: true},
		Data: &dto.TransferStatus{
			TransferID:      "transfer-123",
			Status:          "validated",
			ReferenceNumber: "ref-123",
			Amount:          decimal.NewFromFloat(500.00),
			Currency:        "USD",
		},
	}

	respBody, _ := json.Marshal(validationResponse)

	originalClient := suite.service.client
	suite.service.client = &http.Client{
		Transport: &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(respBody)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			},
		},
	}
	defer func() { suite.service.client = originalClient }()

	ctx := context.Background()
	result, err := suite.service.ValidateTransfer(ctx, dto.TransferValidationInput{
		ReferenceNumber: "ref-123",
		SourceUser:      sourceUser,
		DestinationUser: destinationUser,
		FromAccount:     fromAccount,
		ToAccount:       toAccount,
		Amount:          decimal.NewFromFloat(500.00),
		Description:     "Payment for services",
	})

	suite.NoError(err)
	suite.NotNil(result)
	suite.True(result.Validation.Valid)
	suite.Equal("transfer-123", result.Data.TransferID)
	suite.Equal("validated", result.Data.Status)
}

func (suite *NorthWindServiceTestSuite) TestValidateTransfer_ValidationFailed() {
	sourceUser := &models.User{
		ID:        uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d483"),
		FirstName: "John",
		LastName:  "Doe",
	}

	destinationUser := &models.User{
		ID:        uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d484"),
		FirstName: "Jane",
		LastName:  "Smith",
	}

	fromAccount := &models.Account{
		ID:            uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d485"),
		AccountNumber: "1000000001",
		RoutingNumber: "021000021",
	}

	toAccount := &models.Account{
		ID:            uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d486"),
		AccountNumber: "1000000002",
		RoutingNumber: "021000021",
	}

	errResponse := dto.ErrorResponse{
		Error: dto.ErrorDetail{
			Code:      "VALIDATION_ERROR",
			Message:   "Transfer amount exceeds daily limit",
			RequestID: "req-123",
		},
	}

	respBody, _ := json.Marshal(errResponse)

	originalClient := suite.service.client
	suite.service.client = &http.Client{
		Transport: &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(bytes.NewReader(respBody)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			},
		},
	}
	defer func() { suite.service.client = originalClient }()

	ctx := context.Background()
	result, err := suite.service.ValidateTransfer(ctx, dto.TransferValidationInput{
		ReferenceNumber: "ref-123",
		SourceUser:      sourceUser,
		DestinationUser: destinationUser,
		FromAccount:     fromAccount,
		ToAccount:       toAccount,
		Amount:          decimal.NewFromFloat(50000.00),
		Description:     "Large payment",
	})

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "Transfer amount exceeds daily limit")
}

// ============================================================================
// InitiateTransfer Tests
// ============================================================================

func (suite *NorthWindServiceTestSuite) TestInitiateTransfer_Success() {
	transferDto := dto.TransferInitiationRequest{
		Amount:          500.00,
		Currency:        "USD",
		ReferenceNumber: "ref-123",
		Description:     "Payment",
		SourceAccount: dto.TransferAccount{
			AccountHolderName: "John Doe",
			AccountNumber:     "1000000001",
			RoutingNumber:     "021000021",
			InstitutionName:   "Acme Bank",
		},
		DestinationAccount: dto.TransferAccount{
			AccountHolderName: "Jane Smith",
			AccountNumber:     "1000000002",
			RoutingNumber:     "021000021",
			InstitutionName:   "Acme Bank",
		},
	}

	transferResponse := dto.TransferStatus{
		TransferID:      "transfer-123",
		Status:          "pending",
		ReferenceNumber: "ref-123",
		Amount:          decimal.NewFromFloat(500.00),
		Currency:        "USD",
	}

	respBody, _ := json.Marshal(transferResponse)

	originalClient := suite.service.client
	suite.service.client = &http.Client{
		Transport: &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusCreated,
					Body:       io.NopCloser(bytes.NewReader(respBody)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			},
		},
	}
	defer func() { suite.service.client = originalClient }()

	result, err := suite.service.InitiateTransfer(context.Background(), transferDto)

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal("transfer-123", result.TransferID)
	suite.Equal("pending", result.Status)
	suite.Equal("ref-123", result.ReferenceNumber)
}

func (suite *NorthWindServiceTestSuite) TestInitiateTransfer_BadRequest() {
	transferDto := dto.TransferInitiationRequest{
		Amount:          -100.00,
		Currency:        "USD",
		ReferenceNumber: "ref-invalid",
	}

	errResponse := dto.ErrorResponse{
		Error: dto.ErrorDetail{
			Code:      "VALIDATION_ERROR",
			Message:   "Amount must be positive",
			RequestID: "req-456",
		},
	}

	respBody, _ := json.Marshal(errResponse)

	originalClient := suite.service.client
	suite.service.client = &http.Client{
		Transport: &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(bytes.NewReader(respBody)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			},
		},
	}
	defer func() { suite.service.client = originalClient }()

	result, err := suite.service.InitiateTransfer(context.Background(), transferDto)

	suite.Error(err)
	suite.Nil(result)
}

func (suite *NorthWindServiceTestSuite) TestInitiateTransfer_Unauthorized() {
	transferDto := dto.TransferInitiationRequest{
		Amount:          500.00,
		Currency:        "USD",
		ReferenceNumber: "ref-123",
	}

	errResponse := dto.ErrorResponse{
		Error: dto.ErrorDetail{
			Code:      "AUTH_ERROR",
			Message:   "Invalid API credentials",
			RequestID: "req-789",
		},
	}

	respBody, _ := json.Marshal(errResponse)

	originalClient := suite.service.client
	suite.service.client = &http.Client{
		Transport: &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusUnauthorized,
					Body:       io.NopCloser(bytes.NewReader(respBody)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			},
		},
	}
	defer func() { suite.service.client = originalClient }()

	result, err := suite.service.InitiateTransfer(context.Background(), transferDto)

	suite.Error(err)
	suite.Nil(result)
}

func (suite *NorthWindServiceTestSuite) TestInitiateTransfer_UnexpectedStatus() {
	transferDto := dto.TransferInitiationRequest{
		Amount:          500.00,
		Currency:        "USD",
		ReferenceNumber: "ref-123",
	}

	originalClient := suite.service.client
	suite.service.client = &http.Client{
		Transport: &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusMovedPermanently,
					Body:       io.NopCloser(bytes.NewReader([]byte("Moved"))),
					Header:     http.Header{"Content-Type": []string{"text/plain"}},
				}, nil
			},
		},
	}
	defer func() { suite.service.client = originalClient }()

	result, err := suite.service.InitiateTransfer(context.Background(), transferDto)

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "northwind error (301)")
}

// ============================================================================
// GetTransferStatus Tests
// ============================================================================

func (suite *NorthWindServiceTestSuite) TestGetTransferStatus_Success() {
	transferResponse := dto.TransferStatus{
		TransferID:      "transfer-123",
		Status:          "completed",
		ReferenceNumber: "ref-123",
		Amount:          decimal.NewFromFloat(500.00),
		Currency:        "USD",
	}

	respBody, _ := json.Marshal(transferResponse)

	originalClient := suite.service.client
	suite.service.client = &http.Client{
		Transport: &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(respBody)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			},
		},
	}
	defer func() { suite.service.client = originalClient }()

	result, err := suite.service.GetTransferStatus(context.Background(), "transfer-123")

	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal("transfer-123", result.TransferID)
	suite.Equal("completed", result.Status)
}

func (suite *NorthWindServiceTestSuite) TestGetTransferStatus_NotFound() {
	originalClient := suite.service.client
	suite.service.client = &http.Client{
		Transport: &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(bytes.NewReader([]byte("Not found"))),
					Header:     http.Header{"Content-Type": []string{"text/plain"}},
				}, nil
			},
		},
	}
	defer func() { suite.service.client = originalClient }()

	result, err := suite.service.GetTransferStatus(context.Background(), "nonexistent-transfer")

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "northwind error (404)")
}

func (suite *NorthWindServiceTestSuite) TestGetTransferStatus_BadRequest() {
	errResponse := dto.ErrorResponse{
		Error: dto.ErrorDetail{
			Code:      "VALIDATION_ERROR",
			Message:   "Invalid transfer ID format",
			RequestID: "req-123",
		},
	}

	respBody, _ := json.Marshal(errResponse)

	originalClient := suite.service.client
	suite.service.client = &http.Client{
		Transport: &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(bytes.NewReader(respBody)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			},
		},
	}
	defer func() { suite.service.client = originalClient }()

	result, err := suite.service.GetTransferStatus(context.Background(), "invalid-id")

	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "Invalid transfer ID format")
}

// ============================================================================
// Notify Tests
// ============================================================================

func (suite *NorthWindServiceTestSuite) TestNotify_Success() {
	transfer := &models.Transfer{
		ID:              uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d487"),
		ReferenceNumber: "ref-123",
		Status:          "completed",
	}

	originalClient := suite.service.client
	suite.service.client = &http.Client{
		Transport: &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte("OK"))),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			},
		},
	}
	defer func() { suite.service.client = originalClient }()

	err := suite.service.Notify(context.TODO(), map[string]interface{}{
		"transfer_id": transfer.ID.String(),
		"status":      transfer.Status,
	})

	suite.NoError(err)
}

// ============================================================================
// Helper Mock Types
// ============================================================================

type mockTransport struct {
	roundTrip func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTrip(req)
}

// ============================================================================
// Test Suite Runner
// ============================================================================

func TestNorthWindServiceSuite(t *testing.T) {
	suite.Run(t, new(NorthWindServiceTestSuite))
}
