package ipquality

import (
	"fmt"
	"strings"
)

func contains(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), sub)
}

func itoa(i uint) string {
	return fmt.Sprintf("%d", i)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func merge(a, b *IPInfo) *IPInfo {
	if b == nil {
		return a
	}

	if b.Org != "" {
		a.Org = b.Org
	}

	a.IsHosting = b.IsHosting

	return a
}
