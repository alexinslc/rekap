package collectors

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// NetworkResult contains network usage information
type NetworkResult struct {
	InterfaceName    string
	NetworkName      string // WiFi SSID or "Ethernet"
	BytesReceived    int64
	BytesSent        int64
	Available        bool
	Error            error
}

// CollectNetwork retrieves current network usage statistics
// This collects data from the network interface statistics
func CollectNetwork(ctx context.Context) NetworkResult {
	result := NetworkResult{Available: false}

	// Get active network interface
	iface, ifaceType, err := getActiveInterface(ctx)
	if err != nil {
		result.Error = fmt.Errorf("failed to get active interface: %w", err)
		return result
	}

	result.InterfaceName = iface

	// Get WiFi SSID if on WiFi
	if ifaceType == "WiFi" {
		ssid, err := getWiFiSSID(ctx, iface)
		if err == nil && ssid != "" {
			result.NetworkName = ssid
		} else {
			result.NetworkName = "WiFi"
		}
	} else {
		result.NetworkName = ifaceType
	}

	// Get network statistics for the interface
	bytesRecv, bytesSent, err := getInterfaceStats(ctx, iface)
	if err != nil {
		result.Error = fmt.Errorf("failed to get interface stats: %w", err)
		return result
	}

	result.BytesReceived = bytesRecv
	result.BytesSent = bytesSent
	result.Available = true

	return result
}

// getActiveInterface returns the active network interface name and type
func getActiveInterface(ctx context.Context) (string, string, error) {
	// Use route get to find the interface for default route
	cmd := exec.CommandContext(ctx, "route", "-n", "get", "default")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("route command failed: %w", err)
	}

	// Parse output to find interface
	// Output looks like: "interface: en0"
	re := regexp.MustCompile(`interface:\s*(\w+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return "", "", fmt.Errorf("failed to parse interface from route output")
	}

	iface := matches[1]

	// Determine interface type based on name
	ifaceType := "Ethernet"
	if strings.HasPrefix(iface, "en") {
		// Check if it's WiFi (en0 is typically WiFi on Mac)
		cmd := exec.CommandContext(ctx, "networksetup", "-listallhardwareports")
		output, err := cmd.Output()
		if err == nil {
			// Parse to check if this interface is WiFi
			if strings.Contains(string(output), "Wi-Fi") && strings.Contains(string(output), iface) {
				ifaceType = "WiFi"
			}
		}
	} else if strings.HasPrefix(iface, "bridge") {
		ifaceType = "Bridge"
	} else if strings.HasPrefix(iface, "utun") || strings.HasPrefix(iface, "ipsec") {
		ifaceType = "VPN"
	}

	return iface, ifaceType, nil
}

// getWiFiSSID returns the current WiFi SSID for the given interface
func getWiFiSSID(ctx context.Context, iface string) (string, error) {
	// Use airport command to get SSID (undocumented but reliable)
	// Note: This is a private framework path and may change in future macOS versions
	airportPath := "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport"
	cmd := exec.CommandContext(ctx, airportPath, "-I")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to networksetup with dynamic interface name
		cmd = exec.CommandContext(ctx, "networksetup", "-getairportnetwork", iface)
		output, err = cmd.Output()
		if err != nil {
			return "", err
		}
		// Parse "Current Wi-Fi Network: NetworkName"
		parts := strings.Split(string(output), ":")
		if len(parts) >= 2 {
			return strings.TrimSpace(parts[1]), nil
		}
		return "", fmt.Errorf("failed to parse SSID")
	}

	// Parse output like: " SSID: NetworkName"
	re := regexp.MustCompile(`\s*SSID:\s*(.+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1]), nil
	}

	return "", fmt.Errorf("SSID not found in airport output")
}

// getInterfaceStats returns bytes received and sent for an interface
func getInterfaceStats(ctx context.Context, iface string) (int64, int64, error) {
	// Use netstat to get interface statistics
	cmd := exec.CommandContext(ctx, "netstat", "-ib", "-I", iface)
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("netstat command failed: %w", err)
	}

	// Parse netstat output
	// Expected format: Name Mtu Network Address Ipkts Ierrs Ibytes Opkts Oerrs Obytes Coll
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0, 0, fmt.Errorf("unexpected netstat output format")
	}

	// Find header line to determine field positions
	const (
		fieldIbytes = 6 // Default position for Ibytes
		fieldObytes = 9 // Default position for Obytes
	)
	
	headerLine := lines[0]
	headerFields := strings.Fields(headerLine)
	
	// Try to find the actual positions of Ibytes and Obytes in case format changes
	ibytesIdx := fieldIbytes
	obytesIdx := fieldObytes
	for i, field := range headerFields {
		if field == "Ibytes" {
			ibytesIdx = i
		} else if field == "Obytes" {
			obytesIdx = i
		}
	}

	// Find the line with the interface stats (not Link#)
	var statsLine string
	for _, line := range lines[1:] {
		if strings.HasPrefix(line, iface) && !strings.Contains(line, "Link#") {
			statsLine = line
			break
		}
	}

	if statsLine == "" {
		// If no non-Link line found, use the first interface line
		for _, line := range lines[1:] {
			if strings.HasPrefix(line, iface) {
				statsLine = line
				break
			}
		}
	}

	if statsLine == "" {
		return 0, 0, fmt.Errorf("no stats found for interface %s", iface)
	}

	// Split by whitespace and extract bytes
	fields := strings.Fields(statsLine)
	minFields := obytesIdx + 1
	if len(fields) < minFields {
		return 0, 0, fmt.Errorf("unexpected number of fields in netstat output: %d (expected at least %d)", len(fields), minFields)
	}

	// Parse bytes using the determined field indices
	bytesRecv, err := strconv.ParseInt(fields[ibytesIdx], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse bytes received: %w", err)
	}

	bytesSent, err := strconv.ParseInt(fields[obytesIdx], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse bytes sent: %w", err)
	}

	return bytesRecv, bytesSent, nil
}

// FormatBytes formats bytes into human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
