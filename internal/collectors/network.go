package collectors

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// NetworkResult contains network usage information
type NetworkResult struct {
	InterfaceName string
	NetworkName   string // WiFi SSID or "Ethernet"
	BytesReceived int64
	BytesSent     int64
	SinceBoot     bool // true if stats are since boot (no baseline available)
	Available     bool
	Error         error
}

// networkBaseline stores the first-of-day network stats for delta calculation
type networkBaseline struct {
	Interface     string `json:"interface"`
	BytesReceived int64  `json:"bytes_received"`
	BytesSent     int64  `json:"bytes_sent"`
	Timestamp     string `json:"timestamp"`
}

// CollectNetwork retrieves current network usage statistics
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

	result.Available = true

	// Try to compute today-only delta from baseline
	baseline, err := loadNetworkBaseline()
	if err != nil || baseline.Interface != iface {
		// No baseline or different interface -- save current as baseline, show since-boot
		_ = saveNetworkBaseline(iface, bytesRecv, bytesSent)
		result.BytesReceived = bytesRecv
		result.BytesSent = bytesSent
		result.SinceBoot = true
		return result
	}

	// Compute delta. If current < baseline, counters reset (reboot) -- use current as-is
	recvDelta := bytesRecv - baseline.BytesReceived
	sentDelta := bytesSent - baseline.BytesSent
	if recvDelta < 0 || sentDelta < 0 {
		// Counter reset (reboot). Save new baseline, show current values.
		_ = saveNetworkBaseline(iface, bytesRecv, bytesSent)
		result.BytesReceived = bytesRecv
		result.BytesSent = bytesSent
		result.SinceBoot = true
		return result
	}

	result.BytesReceived = recvDelta
	result.BytesSent = sentDelta
	result.SinceBoot = false

	return result
}

func baselinePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	date := time.Now().Format("2006-01-02")
	return filepath.Join(homeDir, ".local", "share", "rekap", fmt.Sprintf("network-%s.json", date))
}

func loadNetworkBaseline() (networkBaseline, error) {
	path := baselinePath()
	if path == "" {
		return networkBaseline{}, fmt.Errorf("no home directory")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return networkBaseline{}, err
	}

	var b networkBaseline
	if err := json.Unmarshal(data, &b); err != nil {
		return networkBaseline{}, err
	}
	return b, nil
}

func saveNetworkBaseline(iface string, bytesRecv, bytesSent int64) error {
	path := baselinePath()
	if path == "" {
		return fmt.Errorf("no home directory")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	b := networkBaseline{
		Interface:     iface,
		BytesReceived: bytesRecv,
		BytesSent:     bytesSent,
		Timestamp:     time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(b)
	if err != nil {
		return err
	}

	// Atomic write: write to temp file, then rename into place
	tmpFile, err := os.CreateTemp(dir, "network-baseline-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return err
	}

	// Clean up old baseline files (older than 7 days)
	cleanOldBaselines(dir)

	return nil
}

func cleanOldBaselines(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	cutoff := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "network-") && strings.HasSuffix(name, ".json") {
			// Extract date from "network-YYYY-MM-DD.json"
			date := strings.TrimPrefix(name, "network-")
			date = strings.TrimSuffix(date, ".json")
			if len(date) == 10 && date < cutoff {
				os.Remove(filepath.Join(dir, name))
			}
		}
	}
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
	re := regexp.MustCompile(`interface:\s*(\w+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return "", "", fmt.Errorf("failed to parse interface from route output")
	}

	iface := matches[1]

	// Determine interface type based on name
	ifaceType := "Ethernet"
	if strings.HasPrefix(iface, "en") {
		cmd := exec.CommandContext(ctx, "networksetup", "-listallhardwareports")
		output, err := cmd.Output()
		if err == nil {
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
	airportPath := "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport"
	cmd := exec.CommandContext(ctx, airportPath, "-I")
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.CommandContext(ctx, "networksetup", "-getairportnetwork", iface)
		output, err = cmd.Output()
		if err != nil {
			return "", err
		}
		parts := strings.Split(string(output), ":")
		if len(parts) >= 2 {
			return strings.TrimSpace(parts[1]), nil
		}
		return "", fmt.Errorf("failed to parse SSID")
	}

	re := regexp.MustCompile(`\s*SSID:\s*(.+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1]), nil
	}

	return "", fmt.Errorf("SSID not found in airport output")
}

// getInterfaceStats returns bytes received and sent for an interface
func getInterfaceStats(ctx context.Context, iface string) (int64, int64, error) {
	cmd := exec.CommandContext(ctx, "netstat", "-ib", "-I", iface)
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("netstat command failed: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0, 0, fmt.Errorf("unexpected netstat output format")
	}

	const (
		fieldIbytes = 6
		fieldObytes = 9
	)

	headerLine := lines[0]
	headerFields := strings.Fields(headerLine)

	ibytesIdx := fieldIbytes
	obytesIdx := fieldObytes
	for i, field := range headerFields {
		switch field {
		case "Ibytes":
			ibytesIdx = i
		case "Obytes":
			obytesIdx = i
		}
	}

	var statsLine string
	for _, line := range lines[1:] {
		if strings.HasPrefix(line, iface) && !strings.Contains(line, "Link#") {
			statsLine = line
			break
		}
	}

	if statsLine == "" {
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

	fields := strings.Fields(statsLine)
	minFields := obytesIdx + 1
	if len(fields) < minFields {
		return 0, 0, fmt.Errorf("unexpected number of fields in netstat output: %d (expected at least %d)", len(fields), minFields)
	}

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
