package foundation

import (
	"testing"
	"time"

	"github.com/rs/zerolog/log"
)

func TestSemaphore(t *testing.T) {
	t.Run("RunsOnce", func(t *testing.T) {

		semaphore := NewSemaphore(5)

		for i := 0; i < 5; i++ {
			semaphore.Acquire()
			go func(i int) {
				defer semaphore.Release()

				log.Info().Msgf("Running semaphore test goroutine %v...", i)
				time.Sleep(100 * time.Millisecond)
			}(i)
		}

		semaphore.Wait()
	})

	t.Run("ResetsAfterWaitToBeUsedAnotherTime", func(t *testing.T) {

		semaphore := NewSemaphore(5)

		for i := 0; i < 5; i++ {
			semaphore.Acquire()
			go func(i int) {
				defer semaphore.Release()

				log.Info().Msgf("Running first semaphore test goroutine %v...", i)
				time.Sleep(100 * time.Millisecond)
			}(i)
		}

		semaphore.Wait()

		for i := 0; i < 5; i++ {
			semaphore.Acquire()
			go func(i int) {
				defer semaphore.Release()

				log.Info().Msgf("Running second semaphore test goroutine %v...", i)
				time.Sleep(100 * time.Millisecond)
			}(i)
		}

		semaphore.Wait()

	})

	t.Run("RunsWithAcquireInSelect", func(t *testing.T) {

		semaphore := NewSemaphore(5)

		for i := 0; i < 5; i++ {
			select {
			case semaphore.GetAcquireChannel() <- struct{}{}:
				go func(i int) {
					defer semaphore.Release()

					log.Info().Msgf("Running semaphore test goroutine %v...", i)
					time.Sleep(100 * time.Millisecond)
				}(i)

			case <-time.After(1 * time.Second):
				return
			}
		}

		semaphore.Wait()
	})

}
