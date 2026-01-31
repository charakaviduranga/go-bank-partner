package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/array/banking-api/internal/config"
	"github.com/array/banking-api/internal/dto"
	"github.com/shopspring/decimal"
)

type authTransport struct {
	apiKey string
	base   http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return t.base.RoundTrip(req)
}

type NorthWindService struct {
	cfg    *config.NorthWindConfig
	client *http.Client
	logger *slog.Logger
}

func NewNorthWindService(
	cfg *config.NorthWindConfig,
	logger *slog.Logger,
) NorthWindServiceInterface {
	return &NorthWindService{
		cfg: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &authTransport{
				apiKey: cfg.APIKey,
				base:   http.DefaultTransport,
			},
		},
		logger: logger,
	}
}

func (s *NorthWindService) newRequest(
	ctx context.Context,
	method, url string,
	body any,
) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reader = bytes.NewReader(b)
	}

	return http.NewRequestWithContext(ctx, method, url, reader)
}

func (s *NorthWindService) do(req *http.Request) (*http.Response, []byte, error) {
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("northwind request failed",
			"method", req.Method,
			"url", req.URL.String(),
			"error", err,
		)
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read body: %w", err)
	}

	return resp, body, nil
}

// ---------- Account Validation ----------

func (s *NorthWindService) ValidateAccount(
	ctx context.Context,
	req dto.AccountValidationRequest,
) (*dto.AccountValidationResult, error) {

	httpReq, err := s.newRequest(
		ctx,
		http.MethodPost,
		s.cfg.BaseURL+"/external/accounts/validate",
		req,
	)
	if err != nil {
		return nil, err
	}

	resp, body, err := s.do(httpReq)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var result dto.ValidationResponse[dto.Account]
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}

		exists := result.Data != nil && result.Data.ID != ""
		valid := result.Validation.Valid

		s.logger.Info("account validation completed",
			"valid", valid,
			"exists", exists,
			"account_id", result.Data.ID,
		)

		return &dto.AccountValidationResult{
			Response:         &result,
			AccountExists:    exists,
			AccountValid:     valid,
			AvailableBalance: result.Data.AvailableBalance,
		}, nil

	default:
		return nil, parseError(resp.StatusCode, body)
	}
}

// ---------- Transfer Validation
func (s *NorthWindService) ValidateTransfer(
	ctx context.Context,
	in dto.TransferValidationInput,
) (*dto.ValidationResponse[dto.TransferStatus], error) {

	req := dto.TransferValidationRequest{
		Amount:          in.Amount.InexactFloat64(),
		Currency:        "USD",
		Direction:       "INBOUND",
		TransferType:    "ACH",
		Description:     in.Description,
		ReferenceNumber: in.ReferenceNumber,
		ScheduledDate:   time.Now().UTC().Format(time.RFC3339),
		SourceAccount: dto.TransferAccount{
			AccountHolderName: in.SourceUser.FullName(),
			AccountNumber:     in.FromAccount.AccountNumber,
			RoutingNumber:     in.FromAccount.RoutingNumber,
			InstitutionName:   "Acme Bank",
		},
		DestinationAccount: dto.TransferAccount{
			AccountHolderName: in.DestinationUser.FullName(),
			AccountNumber:     in.ToAccount.AccountNumber,
			RoutingNumber:     in.ToAccount.RoutingNumber,
			InstitutionName:   "Acme Bank",
		},
	}

	httpReq, err := s.newRequest(
		ctx,
		http.MethodPost,
		s.cfg.BaseURL+"/external/transfers/validate",
		req,
	)
	if err != nil {
		return nil, err
	}

	resp, body, err := s.do(httpReq)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var result dto.ValidationResponse[dto.TransferStatus]
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}
		return &result, nil

	default:
		return nil, parseError(resp.StatusCode, body)
	}
}

// ---------- Initiate Transfer ----------

func (s *NorthWindService) InitiateTransfer(
	ctx context.Context,
	req dto.TransferInitiationRequest,
) (*dto.TransferStatus, error) {
	s.logger.Info("Transfer Request DTO", "request", req)

	httpReq, err := s.newRequest(
		ctx,
		http.MethodPost,
		s.cfg.BaseURL+"/external/transfers/initiate",
		req,
	)
	if err != nil {
		return nil, err
	}

	resp, body, err := s.do(httpReq)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusCreated:
		var result dto.TransferStatus
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}

		s.logger.Info("transfer initiated",
			"transfer_id", result.TransferID,
			"status", result.Status,
		)

		return &result, nil

	default:
		return nil, parseError(resp.StatusCode, body)
	}
}

// ---------- Transfer Status ----------

func (s *NorthWindService) GetTransferStatus(
	ctx context.Context,
	transferID string,
) (*dto.TransferStatus, error) {

	httpReq, err := s.newRequest(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/external/transfers/%s", s.cfg.BaseURL, transferID),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, body, err := s.do(httpReq)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var result dto.TransferStatus
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}
		return &result, nil

	default:
		return nil, parseError(resp.StatusCode, body)
	}
}

func parseError(status int, body []byte) error {
	var errResp dto.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return fmt.Errorf("northwind error (%d): %s", status, string(body))
	}
	return errors.New(errResp.Error.Message)
}

func (s *NorthWindService) Notify(ctx context.Context, payload map[string]interface{}) error {
	_, err := s.newRequest(
		ctx,
		http.MethodPost,
		s.cfg.WebhookURL,
		payload,
	)

	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	return nil
}

func (s *NorthWindService) GetAccountBalance(ctx context.Context, accountNumber string) (decimal.Decimal, string, error) {
	httpReq, err := s.newRequest(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/external/accounts/%s/balance", s.cfg.BaseURL, accountNumber),
		nil,
	)
	if err != nil {
		return decimal.Zero, "", err
	}

	resp, body, err := s.do(httpReq)
	if err != nil {
		return decimal.Zero, "", err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var result dto.BalanceResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return decimal.Zero, "", err
		}
		balance := result.AvailableBalance
		return balance, result.Currency, nil

	default:
		return decimal.Zero, "", parseError(resp.StatusCode, body)
	}
}
