package service_mocks

// Note: We cannot directly mock interfaces with generic return types.
// mockgen cannot parse generic types like NorthWindValidateResponse[T].
// To use mocks, manually create mocks for NorthWindServiceInterface or
// refactor to avoid generics in the interface.

// For now, we generate mocks for all interfaces except NorthWindServiceInterface.
// Since mockgen doesn't have a way to exclude specific interfaces,
// we use reflection mode with specific interfaces instead of source mode.

//go:generate mockgen -destination=service_mocks.go -package=service_mocks github.com/array/banking-api/internal/services AccountAssociationServiceInterface,AccountServiceInterface,AccountMetricsServiceInterface,AccountSummaryServiceInterface,AuditServiceInterface,AuthServiceInterface,CategoryServiceInterface,CustomerProfileServiceInterface,CustomerSearchServiceInterface,PasswordServiceInterface,TokenServiceInterface,TransactionProcessingServiceInterface,TransactionGeneratorInterface,StatementServiceInterface,AuditLoggerInterface,MetricsRecorderInterface,CircuitBreakerInterface,CustomerLoggerInterface

// This file contains the go:generate directive to generate mocks for service interfaces.
// To regenerate the mocks, run:
//   go generate ./internal/services/service_mocks
