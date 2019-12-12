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

// DelayTypeFunc allows to override the DelayType
type DelayTypeFunc func(n uint, config *RetryConfig) time.Duration

// ExponentialJitterBackoff returns ever increasing backoffs by a power of 2
// with +/- 0-25% to prevent sychronized requests.
func ExponentialJitterBackoff(n uint, config *RetryConfig) time.Duration {
	ms := ApplyJitter(config.DelayMillisecond * int(1<<n))
	return time.Duration(ms) * time.Millisecond
}

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

type RetryConfig struct {
	Attempts         uint
	DelayMillisecond int
	DelayType        DelayTypeFunc
	LastErrorOnly    bool
}

// Retry retries a function
func Retry(retryableFunc func() error, opts ...RetryOption) error {
	var n uint

	//default
	config := &RetryConfig{
		Attempts:         3,
		DelayMillisecond: 100,
		DelayType:        ExponentialJitterBackoff,
		LastErrorOnly:    false,
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
