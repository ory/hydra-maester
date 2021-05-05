package helpers

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateHttpClient(t *testing.T) {
	t.Run("should create insecureSkipVerify client", func(t *testing.T) {
		client := CreateHttpClient(true, "")
		require.NotNil(t, client)
	})

	t.Run("should create client with insecureSkipVerify and wrong tlsTrustStore", func(t *testing.T) {
		tlsTrustStore := "some path"
		client := CreateHttpClient(true, tlsTrustStore)
		require.Nil(t, client)
	})

	t.Run("should create client with and tlsTrustStore", func(t *testing.T) {
		file, err := ioutil.TempFile("/tmp", "test")
		require.Nil(t, err)
		client := CreateHttpClient(true, file.Name())
		defer os.Remove(file.Name())
		require.NotNil(t, client)
	})

	t.Run("should create client without and tlsTrustStore", func(t *testing.T) {
		client := CreateHttpClient(true, "")
		require.NotNil(t, client)
	})
}
