// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/ory/hydra-maester/hydra"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	hydrav1alpha1 "github.com/ory/hydra-maester/api/v1alpha1"
	"github.com/ory/hydra-maester/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = apiv1.AddToScheme(scheme)
	_ = hydrav1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var (
		metricsAddr, hydraURL, endpoint, forwardedProto, syncPeriod, tlsTrustStore, namespace, leaderElectorNs string
		hydraPort                                                                                              int
		enableLeaderElection, insecureSkipVerify                                                               bool
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
	flag.StringVar(&namespace, "namespace", "", "Namespace in which the controller should operate. Setting this will make the controller ignore other namespaces.")
	flag.StringVar(&leaderElectorNs, "leader-elector-namespace", "", "Leader elector namespace where controller should be set.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	syncPeriodParsed, err := time.ParseDuration(syncPeriod)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: metricsAddr,
		},
		LeaderElection: enableLeaderElection,
		Cache: cache.Options{
			SyncPeriod: &syncPeriodParsed,
			DefaultNamespaces: map[string]cache.Config{
				namespace: {},
			},
		},
		LeaderElectionNamespace: leaderElectorNs,
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

	hydraClient, err := hydra.New(defaultSpec, tlsTrustStore, insecureSkipVerify)
	if err != nil {
		setupLog.Error(err, "making default hydra client", "controller", "OAuth2Client")
		os.Exit(1)

	}

	err = controllers.New(
		mgr.GetClient(),
		hydraClient,
		ctrl.Log.WithName("controllers").WithName("OAuth2Client"),
		controllers.WithNamespace(namespace),
	).SetupWithManager(mgr)
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
