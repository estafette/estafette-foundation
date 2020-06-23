package foundation

import (
	"fmt"
	"strings"
	"time"
)

// RetryError contains all errors for each failed attempt
type RetryError []error

// Error method return string representation of Error
// It is an implementation of error interface
func (e RetryError) Error() string {
	logWithNumber := make([]string, lenWithoutNil(e))
	for i, l := range e {
		if l != nil {
			logWithNumber[i] = fmt.Sprintf("#%d: %s", i+1, l.Error())
		}
	}

	return fmt.Sprintf("All attempts fail:\n%s", strings.Join(logWithNumber, "\n"))
}

// RetryOption allows to override config
type RetryOption func(*RetryConfig)

// Attempts set count of retry
// default is 3
func Attempts(attempts uint) RetryOption {
	return func(c *RetryConfig) {
		c.Attempts = attempts
	}
}

// DelayMillisecond sets delay between retries
// default is 100ms
func DelayMillisecond(delayMilliSeconds int) RetryOption {
	return func(c *RetryConfig) {
		c.DelayMillisecond = delayMilliSeconds
	}
}

// ExponentialJitterBackoff sets ExponentialJitterBackoffDelay as DelayType
func ExponentialJitterBackoff() RetryOption {
	return func(c *RetryConfig) {
		c.DelayType = ExponentialJitterBackoffDelay
	}
}

// ExponentialBackOff sets ExponentialBackOffDelay as DelayType
func ExponentialBackOff() RetryOption {
	return func(c *RetryConfig) {
		c.DelayType = ExponentialBackOffDelay
	}
}

// Fixed sets FixedDelay as DelayType
func Fixed() RetryOption {
	return func(c *RetryConfig) {
		c.DelayType = FixedDelay
	}
}

// AnyError is sets AnyErrorIsRetryable as IsRetryableError
func AnyError() RetryOption {
	return func(c *RetryConfig) {
		c.IsRetryableError = AnyErrorIsRetryable
	}
}

// DelayTypeFunc allows to override the DelayType
type DelayTypeFunc func(n uint, config *RetryConfig) time.Duration

// ExponentialJitterBackoffDelay returns ever increasing backoffs by a power of 2
// with +/- 0-25% to prevent sychronized requests.
func ExponentialJitterBackoffDelay(n uint, config *RetryConfig) time.Duration {
	ms := ApplyJitter(config.DelayMillisecond * int(1<<n))
	return time.Duration(ms) * time.Millisecond
}

// ExponentialBackOffDelay is a DelayType which increases delay between consecutive retries exponentially
func ExponentialBackOffDelay(n uint, config *RetryConfig) time.Duration {
	return time.Duration(config.DelayMillisecond) * (1 << n)
}

// FixedDelay is a DelayType which keeps delay the same through all iterations
func FixedDelay(_ uint, config *RetryConfig) time.Duration {
	return time.Duration(config.DelayMillisecond)
}

// IsRetryableErrorFunc allows to apply custom logic to whether an error is retryable
type IsRetryableErrorFunc func(err error) bool

// AnyErrorIsRetryable is a IsRetryableErrorFunc which returns whether an error should be retried
func AnyErrorIsRetryable(err error) bool {
	return err != nil
}

// RetryConfig is used to configure the Retry function
type RetryConfig struct {
	Attempts         uint
	DelayMillisecond int
	DelayType        DelayTypeFunc
	LastErrorOnly    bool
	IsRetryableError IsRetryableErrorFunc
}

// Retry retries a function
func Retry(retryableFunc func() error, opts ...RetryOption) error {
	var n uint

	//default
	config := &RetryConfig{
		Attempts:         3,
		DelayMillisecond: 100,
		DelayType:        ExponentialJitterBackoffDelay,
		LastErrorOnly:    false,
		IsRetryableError: AnyErrorIsRetryable,
	}

	// apply options to override config defaults
	for _, opt := range opts {
		opt(config)
	}

	var errorLog RetryError
	if !config.LastErrorOnly {
		errorLog = make(RetryError, config.Attempts)
	} else {
		errorLog = make(RetryError, 1)
	}

	lastErrIndex := n
	for n < config.Attempts {
		err := retryableFunc()

		if err != nil {
			errorLog[lastErrIndex] = unpackUnrecoverable(err)

			// if this is last attempt - don't wait
			if n == config.Attempts-1 {
				break
			}

			// if this error shouldn't be retried don't retry
			if !config.IsRetryableError(err) {
				break
			}

			delayTime := config.DelayType(n, config)
			time.Sleep(delayTime)
		} else {
			return nil
		}

		n++
		if !config.LastErrorOnly {
			lastErrIndex = n
		}
	}

	if config.LastErrorOnly {
		return errorLog[lastErrIndex]
	}
	return errorLog
}

func unpackUnrecoverable(err error) error {
	if unrecoverable, isUnrecoverable := err.(unrecoverableError); isUnrecoverable {
		return unrecoverable.error
	}

	return err
}

type unrecoverableError struct {
	error
}

func lenWithoutNil(e RetryError) (count int) {
	for _, v := range e {
		if v != nil {
			count++
		}
	}

	return
}
