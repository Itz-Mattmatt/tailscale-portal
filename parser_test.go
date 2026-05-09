package main

import (
	"testing"
)

func TestParseStatus(t *testing.T) {
	jsonData := `{
		"TCP": {
			"3000": { "HTTPS": true },
			"4000": { "HTTPS": true },
			"443": { "HTTPS": true }
		},
		"Web": {
			"personal-macbook.tailf46b52.ts.net:3000": {
				"Handlers": {
					"/": { "Proxy": "http://localhost:5173" }
				}
			},
			"personal-macbook.tailf46b52.ts.net:443": {
				"Handlers": {
					"/": { "Proxy": "http://127.0.0.1:5173" },
					"/opencode": { "Proxy": "http://localhost:4096" }
				}
			}
		}
	}`

	status, err := ParseStatus([]byte(jsonData))
	if err != nil {
		t.Fatalf("Failed to parse status: %v", err)
	}

	// Check TCP configs
	if len(status.TCP) != 3 {
		t.Errorf("Expected 3 TCP ports, got %d", len(status.TCP))
	}

	if !status.TCP["3000"].HTTPS {
		t.Error("Expected port 3000 to have HTTPS enabled")
	}

	// Check Web configs
	if len(status.Web) != 2 {
		t.Errorf("Expected 2 Web services, got %d", len(status.Web))
	}

	service3000 := status.Web["personal-macbook.tailf46b52.ts.net:3000"]
	if len(service3000.Handlers) != 1 {
		t.Errorf("Expected 1 handler for port 3000, got %d", len(service3000.Handlers))
	}

	service443 := status.Web["personal-macbook.tailf46b52.ts.net:443"]
	if len(service443.Handlers) != 2 {
		t.Errorf("Expected 2 handlers for port 443, got %d", len(service443.Handlers))
	}
}

func TestServicesFromStatus(t *testing.T) {
	status := TailscaleStatus{
		TCP: map[string]TCPConfig{
			"3000": {HTTPS: true},
			"443":  {HTTPS: true},
		},
		Web: map[string]WebConfig{
			"personal-macbook.tailf46b52.ts.net:3000": {
				Handlers: map[string]HandlerConfig{
					"/": {Proxy: "http://localhost:5173"},
				},
			},
			"personal-macbook.tailf46b52.ts.net:443": {
				Handlers: map[string]HandlerConfig{
					"/":         {Proxy: "http://127.0.0.1:5173"},
					"/opencode": {Proxy: "http://localhost:4096"},
				},
			},
		},
	}

	services := ServicesFromStatus(status)

	// Should have 3 services total (1 for port 3000, 2 for port 443)
	if len(services) != 3 {
		t.Errorf("Expected 3 services, got %d", len(services))
	}

	// Check first service
	found := false
	for _, svc := range services {
		if svc.Port == "3000" && svc.Path == "/" {
			found = true
			if svc.Protocol != "HTTPS" {
				t.Error("Expected HTTPS protocol for port 3000")
			}
			if svc.Target != "http://localhost:5173" {
				t.Errorf("Expected target http://localhost:5173, got %s", svc.Target)
			}
			if svc.FullURL != "https://personal-macbook.tailf46b52.ts.net:3000/" {
				t.Errorf("Expected URL https://personal-macbook.tailf46b52.ts.net:3000/, got %s", svc.FullURL)
			}
		}
	}
	if !found {
		t.Error("Could not find service for port 3000")
	}
}

func TestBuildFullURL(t *testing.T) {
	tests := []struct {
		hostname string
		port     string
		protocol string
		path     string
		expected string
	}{
		{
			hostname: "example.ts.net",
			port:     "443",
			protocol: "HTTPS",
			path:     "/",
			expected: "https://example.ts.net:443/",
		},
		{
			hostname: "example.ts.net",
			port:     "3000",
			protocol: "HTTP",
			path:     "/api",
			expected: "http://example.ts.net:3000/api",
		},
		{
			hostname: "example.ts.net",
			port:     "8080",
			protocol: "HTTPS",
			path:     "dashboard",
			expected: "https://example.ts.net:8080/dashboard",
		},
	}

	for _, tt := range tests {
		result := buildFullURL(tt.hostname, tt.port, tt.protocol, tt.path)
		if result != tt.expected {
			t.Errorf("buildFullURL(%s, %s, %s, %s) = %s, expected %s",
				tt.hostname, tt.port, tt.protocol, tt.path, result, tt.expected)
		}
	}
}
