package main

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/pflag"

	"github.com/furadx/iam-go/internal/apiserver/options"
)

func TestStoreCloseHasSingleOwner(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	if got := strings.Count(string(source), "store.Close()"); got != 1 {
		t.Fatalf("expected exactly one store.Close() owner, got %d", got)
	}
}

func TestApplyFlagOverridesAfterConfig(t *testing.T) {
	opts := options.NewOptions()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	opts.AddFlags(fs)
	if err := fs.Parse([]string{
		"--server.addr", ":18080",
		"--security.rate-limit.api-limit", "10",
		"--security.cors.allowed-origins", "http://cli.example.com",
	}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}
	overrides, err := captureFlagOverrides(fs)
	if err != nil {
		t.Fatalf("capture flag overrides: %v", err)
	}

	// Simulate config file values being unmarshaled after flags were parsed.
	opts.Server.Addr = ":8080"
	opts.Security.RateLimit.APILimit = 300
	opts.Security.CORS.AllowedOrigins = []string{"http://config.example.com"}

	if err := applyFlagOverrides(opts, overrides); err != nil {
		t.Fatalf("apply flag overrides: %v", err)
	}
	if opts.Server.Addr != ":18080" {
		t.Fatalf("expected server addr from flag, got %q", opts.Server.Addr)
	}
	if opts.Security.RateLimit.APILimit != 10 {
		t.Fatalf("expected api limit from flag, got %d", opts.Security.RateLimit.APILimit)
	}
	wantOrigins := []string{"http://cli.example.com"}
	if !reflect.DeepEqual(opts.Security.CORS.AllowedOrigins, wantOrigins) {
		t.Fatalf("expected CORS origins from flag, got %#v", opts.Security.CORS.AllowedOrigins)
	}
}
