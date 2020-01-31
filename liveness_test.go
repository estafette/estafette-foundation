package foundation

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitLiveness(t *testing.T) {

	t.Run("Returns200OK", func(t *testing.T) {

		// act
		InitLiveness()

		resp, err := http.Get("http://localhost:5000/liveness")

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
