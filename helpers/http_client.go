package helpers

import (
	"crypto/tls"
	"net/http"

	ctrl "sigs.k8s.io/controller-runtime"

	httptransport "github.com/go-openapi/runtime/client"
)

func CreateHttpClient(insecureSkipVerify bool, tlsTrustStore string) *http.Client {
	setupLog := ctrl.Log.WithName("setup")
	tr := &http.Transport{}
	httpClient := &http.Client{}
	if insecureSkipVerify {
		setupLog.Info("configuring TLS with InsecureSkipVerify")
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient.Transport = tr
	}
	if tlsTrustStore != "" {
		setupLog.Info("configuring TLS with tlsTrustStore")
		ops := httptransport.TLSClientOptions{
			CA:                 tlsTrustStore,
			InsecureSkipVerify: insecureSkipVerify,
		}
		if tlsClient, err := httptransport.TLSClient(ops); err != nil {
			setupLog.Error(err, "Error while getting TLSClient, default http client will be used")
			return tlsClient
		}
	}
	return httpClient
}
