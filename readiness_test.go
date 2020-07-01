package foundation

import (
	"io/ioutil"
	"testing"

	"github.com/sethgrid/pester"
	"github.com/stretchr/testify/assert"
)

func TestInitReadiness(t *testing.T) {

	t.Run("Returns200OK", func(t *testing.T) {

		// act
		InitReadinessWithPort(5001)

		resp, err := pester.Get("http://localhost:5001/readiness")

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
