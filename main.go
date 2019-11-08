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
	"net/http"
	"net/url"
	"os"

	"github.com/ory/hydra-maester/hydra"

	hydrav1alpha3 "github.com/ory/hydra-maester/api/v1alpha3"
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
	hydrav1alpha3.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var (
		metricsAddr, hydraURL, endpoint, forwardedProto string
		hydraPort                                       int
		enableLeaderElection                            bool
	)

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&hydraURL, "hydra-url", "", "The address of ORY Hydra")
	flag.IntVar(&hydraPort, "hydra-port", 4445, "Port ORY Hydra is listening on")
	flag.StringVar(&endpoint, "endpoint", "/clients", "ORY Hydra's client endpoint")
	flag.StringVar(&forwardedProto, "forwarded-proto", "", "If set, this adds the value as the X-Forwarded-Proto header in requests to the ORY Hydra admin server")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.Logger(true))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if hydraURL == "" {
		setupLog.Error(fmt.Errorf("hydra URL can't be empty"), "unable to create controller", "controller", "OAuth2Client")
		os.Exit(1)
	}

	defaultSpec := hydrav1alpha3.OAuth2ClientSpec{
		HydraURL:            hydraURL,
		HydraPort:           hydraPort,
		HydraEndpoint:       endpoint,
		HydraForwardedProto: forwardedProto,
	}
	hydraClientMaker := getHydraClientMaker(defaultSpec)
	hydraClient, err := hydraClientMaker(defaultSpec)
	if err != nil {
		setupLog.Error(err, "making default hydra client", "controller", "OAuth2Client")
		os.Exit(1)

	}

	err = (&controllers.OAuth2ClientReconciler{
		Client:           mgr.GetClient(),
		Log:              ctrl.Log.WithName("controllers").WithName("OAuth2Client"),
		HydraClient:      hydraClient,
		HydraClientMaker: hydraClientMaker,
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

func getHydraClientMaker(defaultSpec hydrav1alpha3.OAuth2ClientSpec) controllers.HydraClientMakerFunc {

	return controllers.HydraClientMakerFunc(func(spec hydrav1alpha3.OAuth2ClientSpec) (controllers.HydraClientInterface, error) {

		if spec.HydraURL == "" {
			spec.HydraURL = defaultSpec.HydraURL
		}
		if spec.HydraPort == 0 {
			spec.HydraPort = defaultSpec.HydraPort
		}
		if spec.HydraEndpoint == "" {
			spec.HydraEndpoint = defaultSpec.HydraEndpoint
		}
		if spec.HydraForwardedProto == "" {
			spec.HydraForwardedProto = defaultSpec.HydraForwardedProto
		}

		address := fmt.Sprintf("%s:%d", spec.HydraURL, spec.HydraPort)
		u, err := url.Parse(address)
		if err != nil {
			return nil, fmt.Errorf("unable to parse ORY Hydra's URL: %w", err)
		}
		return &hydra.Client{
			HydraURL:       *u.ResolveReference(&url.URL{Path: spec.HydraEndpoint}),
			HTTPClient:     &http.Client{},
			ForwardedProto: spec.HydraForwardedProto,
		}, nil
	})

}
