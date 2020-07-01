package foundation

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitLivenessAndReadiness(t *testing.T) {

	t.Run("Returns200OKForLiveness", func(t *testing.T) {

		// act
		InitLivenessAndReadinessWithPort(5002)

		resp, err := http.Get("http://localhost:5002/liveness")

		if assert.Nil(t, err) {

			assert.Equal(t, 200, resp.StatusCode)

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)

			if assert.Nil(t, err) {
				assert.Equal(t, "I'm alive!\n", string(body))
			}
		}
	})

	t.Run("Returns200OKForReadiness", func(t *testing.T) {

		// act
		InitLivenessAndReadinessWithPort(5003)

		resp, err := http.Get("http://localhost:5003/readiness")

		if assert.Nil(t, err) {

			assert.Equal(t, 200, resp.StatusCode)

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)

			if assert.Nil(t, err) {
				assert.Equal(t, "I'm ready!\n", string(body))
			}
		}
	})
}
