package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseStatus parses the JSON output from tailscale serve status --json
func ParseStatus(data []byte) (TailscaleStatus, error) {
	var status TailscaleStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return TailscaleStatus{}, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return status, nil
}

// ServicesFromStatus converts TailscaleStatus to a flattened list of Services
func ServicesFromStatus(status TailscaleStatus) []Service {
	var services []Service
	services = append(services, flattenWeb(status, false)...)
	for _, fg := range status.Foreground {
		services = append(services, flattenWeb(fg, true)...)
	}
	return services
}

// flattenWeb extracts services from the Web section of a status block.
// isForeground marks services that belong to a foreground process.
func flattenWeb(status TailscaleStatus, isForeground bool) []Service {
	var services []Service

	for serviceName, webConfig := range status.Web {
		// Parse serviceName (hostname:port)
		parts := strings.Split(serviceName, ":")
		if len(parts) != 2 {
			continue // Skip malformed entries
		}
		hostname := parts[0]
		port := parts[1]

		// Determine protocol from TCP section
		protocol := "HTTP"
		if tcpConfig, exists := status.TCP[port]; exists && tcpConfig.HTTPS {
			protocol = "HTTPS"
		}

		// Create a service for each handler
		for path, handler := range webConfig.Handlers {
			fullURL := buildFullURL(hostname, port, protocol, path)

			service := Service{
				ServiceName:  serviceName,
				Hostname:     hostname,
				Port:         port,
				Protocol:     protocol,
				Target:       handler.Proxy,
				Path:         path,
				FullURL:      fullURL,
				IsForeground: isForeground,
			}
			services = append(services, service)
		}
	}

	return services
}

// buildFullURL constructs the full access URL for a service
func buildFullURL(hostname, port, protocol, path string) string {
	scheme := "http"
	if protocol == "HTTPS" {
		scheme = "https"
	}

	// Build URL
	url := fmt.Sprintf("%s://%s:%s", scheme, hostname, port)
	
	// Add path (ensure it starts with /)
	if path != "" && !strings.HasPrefix(path, "/") {
		url += "/" + path
	} else {
		url += path
	}

	return url
}
