package shared

import (
	"errors"
	"fmt"
	"time"
)

func Now() time.Time {
	return time.Now().UTC()
}

var (
	ErrNotFound             = errors.New("resource not found")
	ErrForbidden            = errors.New("access forbidden")
	ErrConflict             = errors.New("resource conflict")
	ErrInvalidTransition    = errors.New("invalid state transition")
	ErrInsufficientStock    = errors.New("insufficient stock available")
	ErrIdempotencyConflict  = errors.New("idempotency key reused with different payload")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrValidationFailed     = errors.New("validation failed")
	ErrPaymentLimitExceeded = errors.New("payment exceeds outstanding amount")
	ErrRefundLimitExceeded  = errors.New("refund exceeds refundable amount")
	ErrUdhaarLimitExceeded  = errors.New("repayment exceeds outstanding udhaar")
	ErrDuplicateKey         = errors.New("duplicate key")
	ErrOptimisticLock       = errors.New("optimistic lock failure")
)

type AppError struct {
	Code    string
	Message string
	Err     error
	Details map[string]string
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

func NewNotFound(resource string) *AppError {
	return &AppError{Code: "NOT_FOUND", Message: resource + " not found", Err: ErrNotFound}
}

func NewForbidden(action string) *AppError {
	return &AppError{Code: "FORBIDDEN", Message: "not authorized to " + action, Err: ErrForbidden}
}

func NewConflict(message string) *AppError {
	return &AppError{Code: "CONFLICT", Message: message, Err: ErrConflict}
}

func NewInvalidTransition(from, to string) *AppError {
	return &AppError{Code: "INVALID_TRANSITION", Message: "cannot transition from " + from + " to " + to, Err: ErrInvalidTransition}
}

func NewInsufficientStock(available, requested int64) *AppError {
	return &AppError{
		Code:    "INSUFFICIENT_STOCK",
		Message: "insufficient stock available",
		Err:     ErrInsufficientStock,
		Details: map[string]string{"available": fmt.Sprintf("%d", available), "requested": fmt.Sprintf("%d", requested)},
	}
}

func NewIdempotencyConflict() *AppError {
	return &AppError{Code: "IDEMPOTENCY_CONFLICT", Message: "idempotency key already used with different payload", Err: ErrIdempotencyConflict}
}

func NewOptimisticLock() *AppError {
	return &AppError{Code: "OPTIMISTIC_LOCK", Message: "concurrent modification, please retry", Err: ErrOptimisticLock}
}

func NewValidationError(field, reason string) *AppError {
	return &AppError{
		Code:    "VALIDATION_ERROR",
		Message: "validation failed",
		Err:     ErrValidationFailed,
		Details: map[string]string{field: reason},
	}
}

func NewPaymentLimitExceeded(outstanding, requested int64) *AppError {
	return &AppError{
		Code:    "PAYMENT_LIMIT_EXCEEDED",
		Message: "payment exceeds outstanding amount",
		Err:     ErrPaymentLimitExceeded,
		Details: map[string]string{"outstanding": fmt.Sprintf("%d", outstanding), "requested": fmt.Sprintf("%d", requested)},
	}
}

func NewRefundLimitExceeded(refundable, requested int64) *AppError {
	return &AppError{
		Code:    "REFUND_LIMIT_EXCEEDED",
		Message: "refund exceeds refundable amount",
		Err:     ErrRefundLimitExceeded,
		Details: map[string]string{"refundable": fmt.Sprintf("%d", refundable), "requested": fmt.Sprintf("%d", requested)},
	}
}

func NewUdhaarLimitExceeded(outstanding, requested int64) *AppError {
	return &AppError{
		Code:    "UDHAAR_LIMIT_EXCEEDED",
		Message: "repayment exceeds outstanding udhaar",
		Err:     ErrUdhaarLimitExceeded,
		Details: map[string]string{"outstanding": fmt.Sprintf("%d", outstanding), "requested": fmt.Sprintf("%d", requested)},
	}
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

func IsInvalidTransition(err error) bool {
	return errors.Is(err, ErrInvalidTransition)
}

func IsInsufficientStock(err error) bool {
	return errors.Is(err, ErrInsufficientStock)
}

func IsIdempotencyConflict(err error) bool {
	return errors.Is(err, ErrIdempotencyConflict)
}

func IsOptimisticLock(err error) bool {
	return errors.Is(err, ErrOptimisticLock)
}

func IsValidationError(err error) bool {
	return errors.Is(err, ErrValidationFailed)
}

func IsPaymentLimitExceeded(err error) bool {
	return errors.Is(err, ErrPaymentLimitExceeded)
}

func IsRefundLimitExceeded(err error) bool {
	return errors.Is(err, ErrRefundLimitExceeded)
}

func IsUdhaarLimitExceeded(err error) bool {
	return errors.Is(err, ErrUdhaarLimitExceeded)
}
