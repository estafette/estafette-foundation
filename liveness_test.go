package foundation

import (
	"io/ioutil"
	"testing"

	"github.com/sethgrid/pester"
	"github.com/stretchr/testify/assert"
)

func TestInitLiveness(t *testing.T) {

	t.Run("Returns200OK", func(t *testing.T) {

		// act
		InitLivenessWithPort(5000)

		resp, err := pester.Get("http://localhost:5000/liveness")

		if assert.Nil(t, err) {

			assert.Equal(t, 200, resp.StatusCode)

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)

			if assert.Nil(t, err) {
				assert.Equal(t, "I'm alive!\n", string(body))
			}
		}
	})
}
