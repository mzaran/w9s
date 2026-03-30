package errs_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mzaran/w9s/internal/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestW9sErrorImplementsError(t *testing.T) {
	var _ error = &errs.W9sError{}
}

func TestErrorMessageWithoutWrapped(t *testing.T) {
	e := &errs.W9sError{
		Code:    errs.ErrConnection,
		Message: "cannot reach server",
	}
	assert.Equal(t, "cannot reach server", e.Error())
}

func TestErrorMessageWithWrapped(t *testing.T) {
	inner := fmt.Errorf("dial tcp: connection refused")
	e := &errs.W9sError{
		Code:    errs.ErrConnection,
		Message: "cannot reach server",
		Err:     inner,
	}
	assert.Equal(t, "cannot reach server: dial tcp: connection refused", e.Error())
}

func TestUnwrap(t *testing.T) {
	inner := fmt.Errorf("original error")
	e := &errs.W9sError{
		Code:    errs.ErrAuth,
		Message: "auth failed",
		Err:     inner,
	}
	assert.Equal(t, inner, e.Unwrap())
}

func TestUnwrapNil(t *testing.T) {
	e := &errs.W9sError{
		Code:    errs.ErrAuth,
		Message: "auth failed",
	}
	assert.Nil(t, e.Unwrap())
}

func TestErrorsIs(t *testing.T) {
	inner := fmt.Errorf("root cause")
	e := errs.Wrap(inner, errs.ErrTimeout, "request timed out")
	assert.True(t, errors.Is(e, inner))
}

func TestWrap(t *testing.T) {
	inner := fmt.Errorf("underlying")
	e := errs.Wrap(inner, errs.ErrValidation, "bad input")

	assert.Equal(t, errs.ErrValidation, e.Code)
	assert.Equal(t, "bad input", e.Message)
	assert.Equal(t, inner, e.Err)
	assert.Contains(t, e.Error(), "underlying")
}

func TestHelperConstructors(t *testing.T) {
	tests := []struct {
		name     string
		ctor     func(error, string) *errs.W9sError
		code     errs.ErrorCode
		message  string
	}{
		{"NewConnectionError", errs.NewConnectionError, errs.ErrConnection, "conn fail"},
		{"NewAuthError", errs.NewAuthError, errs.ErrAuth, "auth fail"},
		{"NewForbiddenError", errs.NewForbiddenError, errs.ErrForbidden, "forbidden"},
		{"NewNotFoundError", errs.NewNotFoundError, errs.ErrNotFound, "not found"},
		{"NewTimeoutError", errs.NewTimeoutError, errs.ErrTimeout, "timeout"},
		{"NewAPIDisabledError", errs.NewAPIDisabledError, errs.ErrAPIDisabled, "api off"},
		{"NewValidationError", errs.NewValidationError, errs.ErrValidation, "invalid"},
		{"NewPowerUnavailableError", errs.NewPowerUnavailableError, errs.ErrPowerUnavailable, "no ipmi"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			inner := fmt.Errorf("cause")
			e := tc.ctor(inner, tc.message)

			require.NotNil(t, e)
			assert.Equal(t, tc.code, e.Code)
			assert.Equal(t, tc.message, e.Message)
			assert.Equal(t, inner, e.Err)
			assert.Contains(t, e.Error(), tc.message)
			assert.Contains(t, e.Error(), "cause")
		})
	}
}

func TestHelperConstructorsNilInner(t *testing.T) {
	e := errs.NewConnectionError(nil, "no server")
	assert.Equal(t, "no server", e.Error())
	assert.Nil(t, e.Unwrap())
}
