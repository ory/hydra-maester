/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/ory/hydra-maester/helpers"

	"github.com/ory/hydra-maester/hydra"

	hydrav1alpha1 "github.com/ory/hydra-maester/api/v1alpha1"
	"github.com/ory/hydra-maester/controllers"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {

	apiv1.AddToScheme(scheme)
	hydrav1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var (
		metricsAddr, hydraURL, endpoint, forwardedProto, syncPeriod, tlsTrustStore string
		hydraPort                                                                  int
		enableLeaderElection, insecureSkipVerify                                   bool
	)

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&hydraURL, "hydra-url", "", "The address of ORY Hydra")
	flag.IntVar(&hydraPort, "hydra-port", 4445, "Port ORY Hydra is listening on")
	flag.StringVar(&endpoint, "endpoint", "/clients", "ORY Hydra's client endpoint")
	flag.StringVar(&forwardedProto, "forwarded-proto", "", "If set, this adds the value as the X-Forwarded-Proto header in requests to the ORY Hydra admin server")
	flag.StringVar(&tlsTrustStore, "tls-trust-store", "", "trust store certificate path. If set ca will be set in http client to connect with hydra admin")
	flag.StringVar(&syncPeriod, "sync-period", "10h", "Determines the minimum frequency at which watched resources are reconciled")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&insecureSkipVerify, "insecure-skip-verify", false, "If set, http client will be configured to skip insecure verification to connect with hydra admin")
	flag.Parse()

	ctrl.SetLogger(zap.Logger(true))

	syncPeriodParsed, err := time.ParseDuration(syncPeriod)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		SyncPeriod:         &syncPeriodParsed,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if hydraURL == "" {
		setupLog.Error(fmt.Errorf("hydra URL can't be empty"), "unable to create controller", "controller", "OAuth2Client")
		os.Exit(1)
	}

	defaultSpec := hydrav1alpha1.OAuth2ClientSpec{
		HydraAdmin: hydrav1alpha1.HydraAdmin{
			URL:            hydraURL,
			Port:           hydraPort,
			Endpoint:       endpoint,
			ForwardedProto: forwardedProto,
		},
	}
	if tlsTrustStore != "" {
		if _, err := os.Stat(tlsTrustStore); err != nil {
			setupLog.Error(err, "cannot parse tls trust store")
			os.Exit(1)
		}
	}

	hydraClient, err := getHydraClient(defaultSpec, tlsTrustStore, insecureSkipVerify)
	if err != nil {
		setupLog.Error(err, "making default hydra client", "controller", "OAuth2Client")
		os.Exit(1)

	}

	err = (&controllers.OAuth2ClientReconciler{
		Client:      mgr.GetClient(),
		Log:         ctrl.Log.WithName("controllers").WithName("OAuth2Client"),
		HydraClient: hydraClient,
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "OAuth2Client")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func getHydraClient(spec hydrav1alpha1.OAuth2ClientSpec, tlsTrustStore string, insecureSkipVerify bool) (controllers.HydraClientInterface, error) {

	address := fmt.Sprintf("%s:%d", spec.HydraAdmin.URL, spec.HydraAdmin.Port)
	u, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	c, err := helpers.CreateHttpClient(insecureSkipVerify, tlsTrustStore)
	if err != nil {
		return nil, err
	}

	client := &hydra.Client{
		HydraURL:   *u.ResolveReference(&url.URL{Path: spec.HydraAdmin.Endpoint}),
		HTTPClient: c,
	}

	if spec.HydraAdmin.ForwardedProto != "" && spec.HydraAdmin.ForwardedProto != "off" {
		client.ForwardedProto = spec.HydraAdmin.ForwardedProto
	}

	return client, nil
}
