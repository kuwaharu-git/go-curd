package main

import (
	"os"
	"testing"
)

func TestTrimSpace(t *testing.T) {
	if got := trimSpace("  hello  "); got != "hello" {
		t.Fatalf("trimSpace() = %q, want %q", got, "hello")
	}
}

func TestGetEnvAsIntFallbackOnInvalid(t *testing.T) {
	key := "TEST_PORT"
	t.Setenv(key, "invalid")

	if got := getEnvAsInt(key, 8080); got != 8080 {
		t.Fatalf("getEnvAsInt() = %d, want fallback %d", got, 8080)
	}
}

func TestGetEnvFallback(t *testing.T) {
	key := "TEST_EMPTY"
	_ = os.Unsetenv(key)

	if got := getEnv(key, "default"); got != "default" {
		t.Fatalf("getEnv() = %q, want %q", got, "default")
	}
}
