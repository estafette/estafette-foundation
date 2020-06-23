package foundation

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	ErrToRetry    = fmt.Errorf("This is an that should be retried")
	ErrToNotRetry = fmt.Errorf("This is an that should NOT be retried")
)

func TestRetry(t *testing.T) {
	t.Run("ReturnsNilIfFunctionSucceeds", func(t *testing.T) {

		attempts := 0
		retryableFunc := func() error {
			attempts++
			return nil
		}

		// act
		err := Retry(retryableFunc)

		assert.Nil(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("ReturnsNilIfFunctionSucceedsBeforeExhaustingAttemps", func(t *testing.T) {

		attempts := 0
		retryableFunc := func() error {
			attempts++
			if attempts < 5 {
				return ErrToRetry
			}
			return nil
		}

		// act
		err := Retry(retryableFunc, Attempts(5), DelayMillisecond(10), Fixed())

		assert.Nil(t, err)
		assert.Equal(t, 5, attempts)
	})

	t.Run("ReturnsErrIfFunctionFailsEveryTime", func(t *testing.T) {

		attempts := 0
		retryableFunc := func() error {
			attempts++
			return ErrToRetry
		}

		// act
		err := Retry(retryableFunc, Attempts(5), DelayMillisecond(10), Fixed())

		assert.NotNil(t, err)
		assert.Equal(t, 5, attempts)
	})

	t.Run("ReturnsLastErrIfFunctionFailsEveryTimeWithLastErrorOnlyTrue", func(t *testing.T) {

		attempts := 0
		retryableFunc := func() error {
			attempts++
			return ErrToRetry
		}

		// act
		err := Retry(retryableFunc, Attempts(5), DelayMillisecond(10), Fixed(), LastErrorOnly(true))

		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, ErrToRetry))
		assert.Equal(t, 5, attempts)
	})

	t.Run("RetriesForCustomRetryableErrorFunc", func(t *testing.T) {

		attempts := 0
		retryableFunc := func() error {
			attempts++
			return ErrToRetry
		}

		isRetryableErrorCustomOption := func(c *RetryConfig) {
			c.IsRetryableError = func(err error) bool {
				if errors.Is(err, ErrToNotRetry) {
					return false
				}
				return true
			}
		}

		// act
		err := Retry(retryableFunc, isRetryableErrorCustomOption, Attempts(5), DelayMillisecond(10), Fixed())

		assert.NotNil(t, err)
		assert.Equal(t, 5, attempts)
	})

	t.Run("DoesNotRetryForCustomRetryableErrorFuncReturningFalse", func(t *testing.T) {

		attempts := 0
		retryableFunc := func() error {
			attempts++
			return ErrToNotRetry
		}

		isRetryableErrorCustomOption := func(c *RetryConfig) {
			c.IsRetryableError = func(err error) bool {
				if errors.Is(err, ErrToNotRetry) {
					return false
				}
				return true
			}
		}

		// act
		err := Retry(retryableFunc, isRetryableErrorCustomOption, Attempts(5), DelayMillisecond(10), Fixed())

		assert.NotNil(t, err)
		assert.Equal(t, 1, attempts)
	})
}
