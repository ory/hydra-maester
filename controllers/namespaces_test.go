// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	hydrav1alpha1 "github.com/ory/hydra-maester/api/v1alpha1"
)

func TestWithNamespaces(t *testing.T) {
	for name, tc := range map[string]struct {
		opts []Option
		want []string
	}{
		"default":        {nil, []string{DefaultNamespace}},
		"single":         {[]Option{WithNamespace("a")}, []string{"a"}},
		"empty is all":   {[]Option{WithNamespace("")}, nil},
		"multiple":       {[]Option{WithNamespaces("a", "b")}, []string{"a", "b"}},
		"skips empty":    {[]Option{WithNamespaces("a", "")}, []string{"a"}},
		"none means all": {[]Option{WithNamespaces()}, nil},
	} {
		t.Run(name, func(t *testing.T) {
			got := New(nil, nil, logr.Discard(), tc.opts...).ControllerNamespaces
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("got %v, want %v", got, tc.want)
				}
			}
		})
	}
}

func TestReconcileNamespaceFilter(t *testing.T) {
	_ = hydrav1alpha1.AddToScheme(scheme.Scheme)

	for name, tc := range map[string]struct {
		watch     []string
		namespace string
		ignored   bool
	}{
		"all namespaces": {nil, "x", false},
		"single match":   {[]string{"a"}, "a", false},
		"single miss":    {[]string{"a"}, "b", true},
		"multi first":    {[]string{"a", "b"}, "a", false},
		"multi second":   {[]string{"a", "b"}, "b", false},
		"multi outside":  {[]string{"a", "b"}, "c", true},
	} {
		t.Run(name, func(t *testing.T) {
			obj := &hydrav1alpha1.OAuth2Client{ObjectMeta: metav1.ObjectMeta{Name: "c", Namespace: tc.namespace}}
			c := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(obj).Build()
			// the hydra client is nil, so only the finalizer is checked to tell filtered from reconciled requests
			r := New(c, nil, logr.Discard(), WithNamespaces(tc.watch...))

			req := ctrl.Request{NamespacedName: types.NamespacedName{Name: "c", Namespace: tc.namespace}}
			func() {
				defer func() { _ = recover() }()
				_, _ = r.Reconcile(context.Background(), req)
			}()

			var got hydrav1alpha1.OAuth2Client
			if err := c.Get(context.Background(), client.ObjectKeyFromObject(obj), &got); err != nil {
				t.Fatal(err)
			}
			if touched := len(got.Finalizers) > 0; touched == tc.ignored {
				t.Fatalf("ignored=%v but finalizers=%v", tc.ignored, got.Finalizers)
			}
		})
	}
}
