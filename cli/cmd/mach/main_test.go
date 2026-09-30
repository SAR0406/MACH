package main

import (
	"testing"

	"github.com/SAR0406/MACH/internal/api"
)

func TestHasTLSEndpoint(t *testing.T) {
	st := api.DeviceStatusResponse{
		Exposures: []api.Exposure{
			{PublicURL: "https://abc.mach.dev"},
		},
	}
	if !hasTLSEndpoint(st) {
		t.Fatal("expected tls endpoint to be detected")
	}

	st.Exposures[0].PublicURL = "http://abc.mach.dev"
	if hasTLSEndpoint(st) {
		t.Fatal("expected non-tls endpoint to fail")
	}
}

func TestHasMACHDNS(t *testing.T) {
	st := api.DeviceStatusResponse{
		Exposures: []api.Exposure{
			{PublicURL: "https://abc.mach.dev"},
		},
	}
	if !hasMACHDNS(st) {
		t.Fatal("expected mach.dev suffix to pass")
	}

	st.Exposures = []api.Exposure{{PublicURL: "https://example.com"}}
	if hasMACHDNS(st) {
		t.Fatal("expected non-mach domain to fail")
	}
}
