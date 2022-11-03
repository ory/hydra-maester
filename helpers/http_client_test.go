// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package helpers_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ory/hydra-maester/helpers"

	"github.com/stretchr/testify/require"
)

func TestCreateHttpClient(t *testing.T) {
	t.Run("should create insecureSkipVerify client", func(t *testing.T) {
		client, err := helpers.CreateHttpClient(true, "")
		require.NotNil(t, client)
		require.Nil(t, err)
	})

	t.Run("should create client with and tlsTrustStore", func(t *testing.T) {
		file, err := ioutil.TempFile("/tmp", "test")
		require.Nil(t, err)
		client, err := helpers.CreateHttpClient(true, file.Name())
		defer os.Remove(file.Name())
		require.NotNil(t, client)
		require.Nil(t, err)
	})

	t.Run("should not create client with and wrong tlsTrustStore", func(t *testing.T) {
		client, err := helpers.CreateHttpClient(true, "/somefile")
		require.Nil(t, client)
		require.NotNil(t, err)
		require.Equal(t, err.Error(), "stat /somefile: no such file or directory")
	})

	t.Run("should create client without and tlsTrustStore", func(t *testing.T) {
		client, err := helpers.CreateHttpClient(true, "")
		require.NotNil(t, client)
		require.Nil(t, err)
	})
}
