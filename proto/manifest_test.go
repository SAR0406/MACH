package proto

import "testing"

func TestParseManifest(t *testing.T) {
	data := []byte(`
name: api
runtime:
  type: container
  image: ghcr.io/acme/api:latest
network:
  http:
    port: 3000
restart:
  policy: always
health:
  command: /health
`)
	m, err := ParseManifest(data)
	if err != nil {
		t.Fatalf("ParseManifest failed: %v", err)
	}
	if m.Network.HTTP.Port != 3000 {
		t.Fatalf("unexpected port: %d", m.Network.HTTP.Port)
	}
}
