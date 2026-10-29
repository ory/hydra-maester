// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package helpers

import "strings"

// SplitNamespaces splits a comma-separated list of namespaces.
// Spaces are trimmed, and empty and duplicate entries are dropped.
func SplitNamespaces(s string) []string {
	var out []string
	seen := map[string]bool{}
	for _, n := range strings.Split(s, ",") {
		n = strings.TrimSpace(n)
		if n != "" && !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	return out
}
