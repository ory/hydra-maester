// Copyright © 2022 Ory Corp

package helpers

import (
	"crypto/tls"
	"net/http"
	"os"

	ctrl "sigs.k8s.io/controller-runtime"

	httptransport "github.com/go-openapi/runtime/client"
)

func CreateHttpClient(insecureSkipVerify bool, tlsTrustStore string) (*http.Client, error) {
	setupLog := ctrl.Log.WithName("setup")
	tr := &http.Transport{}
	httpClient := &http.Client{}
	if insecureSkipVerify {
		setupLog.Info("configuring TLS with InsecureSkipVerify")
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient.Transport = tr
	}
	if tlsTrustStore != "" {
		if _, err := os.Stat(tlsTrustStore); err != nil {
			return nil, err
		}

		setupLog.Info("configuring TLS with tlsTrustStore")
		ops := httptransport.TLSClientOptions{
			CA:                 tlsTrustStore,
			InsecureSkipVerify: insecureSkipVerify,
		}
		if tlsClient, err := httptransport.TLSClient(ops); err != nil {
			setupLog.Error(err, "Error while getting TLSClient, default http client will be used")
			return tlsClient, nil
		}
	}
	return httpClient, nil
}
