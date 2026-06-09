package utils

import (
	"strings"
	"testing"
)

func TestFormatProxy(t *testing.T) {
	tests := map[string]string{
		"":                   "",
		" ":                  "",
		"https://proxy.com":  "https://proxy.com",
		"https://proxy.com/": "https://proxy.com",
	}

	for in, want := range tests {
		if got := formatProxy(in); got != want {
			t.Fatalf("formatProxy(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestJoinProxyURL(t *testing.T) {
	const target = "https://github.com/shaowenchen/ops"

	tests := map[string]string{
		"":                   target,
		"https://proxy.com":  "https://proxy.com/" + target,
		"https://proxy.com/": "https://proxy.com/" + target,
	}

	for proxy, want := range tests {
		if got := joinProxyURL(proxy, target); got != want {
			t.Fatalf("joinProxyURL(%q, %q) = %q, want %q", proxy, target, got, want)
		}
	}
}

func TestGetAvailableURLWithProxy(t *testing.T) {
	const target = "https://github.com/shaowenchen/ops"
	const want = "https://proxy.com/" + target

	for _, proxy := range []string{"https://proxy.com", "https://proxy.com/"} {
		if got := GetAvailableUrl(target, proxy); got != want {
			t.Fatalf("GetAvailableUrl(%q, %q) = %q, want %q", target, proxy, got, want)
		}
	}
}

func TestShellInstallOpscliNormalizesProxy(t *testing.T) {
	script := ShellInstallOpscli("https://proxy.com/")
	if strings.Contains(script, "https://proxy.com//") {
		t.Fatalf("ShellInstallOpscli generated double slash proxy URL: %s", script)
	}
	if !strings.Contains(script, "PROXY=https://proxy.com sh -") {
		t.Fatalf("ShellInstallOpscli should pass normalized PROXY env: %s", script)
	}
}
