// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package helpers

import (
	"reflect"
	"testing"
)

func TestSplitNamespaces(t *testing.T) {
	for in, want := range map[string][]string{
		"":           nil,
		" , ,":       nil,
		"a":          {"a"},
		"a,b":        {"a", "b"},
		" a , b ,,a": {"a", "b"},
	} {
		if got := SplitNamespaces(in); !reflect.DeepEqual(got, want) {
			t.Errorf("SplitNamespaces(%q) = %v, want %v", in, got, want)
		}
	}
}
