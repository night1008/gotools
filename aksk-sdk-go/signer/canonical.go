package signer

import (
	"net/url"
	"sort"
	"strings"
)

func CanonicalQuery(raw string) string {
	if raw == "" {
		return ""
	}

	vals, _ := url.ParseQuery(raw)
	keys := make([]string, 0, len(vals))
	for k := range vals {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		v := vals[k]
		sort.Strings(v)
		for _, item := range v {
			parts = append(parts, k+"="+item)
		}
	}
	return strings.Join(parts, "&")
}
