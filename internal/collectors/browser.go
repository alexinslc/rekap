package collectors

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

// BrowserTab represents a single browser tab
type BrowserTab struct {
	Title  string
	URL    string
	Domain string
}

// BrowserResult contains browser tab information
type BrowserResult struct {
	Tabs      []BrowserTab
	TabCount  int
	Domains   map[string]int // domain -> tab count
	Browser   string
	Available bool
	Error     error
}

// BrowsersResult aggregates all browser data
type BrowsersResult struct {
	Chrome     BrowserResult
	Safari     BrowserResult
	Edge       BrowserResult
	TotalTabs  int
	TopDomains map[string]int // aggregated across all browsers
	Available  bool
}

// CollectBrowserTabs retrieves open tabs from Chrome, Safari, and Edge
func CollectBrowserTabs(ctx context.Context) BrowsersResult {
	result := BrowsersResult{
		TopDomains: make(map[string]int),
	}

	// Collect from each browser concurrently
	chromeChan := make(chan BrowserResult, 1)
	safariChan := make(chan BrowserResult, 1)
	edgeChan := make(chan BrowserResult, 1)

	go func() {
		chromeChan <- collectChromeTabs(ctx)
	}()

	go func() {
		safariChan <- collectSafariTabs(ctx)
	}()

	go func() {
		edgeChan <- collectEdgeTabs(ctx)
	}()

	// Collect results
	result.Chrome = <-chromeChan
	result.Safari = <-safariChan
	result.Edge = <-edgeChan

	// Aggregate data
	result.TotalTabs = result.Chrome.TabCount + result.Safari.TabCount + result.Edge.TabCount

	for domain, count := range result.Chrome.Domains {
		result.TopDomains[domain] += count
	}
	for domain, count := range result.Safari.Domains {
		result.TopDomains[domain] += count
	}
	for domain, count := range result.Edge.Domains {
		result.TopDomains[domain] += count
	}

	result.Available = result.Chrome.Available || result.Safari.Available || result.Edge.Available

	return result
}

func collectChromeTabs(ctx context.Context) BrowserResult {
	result := BrowserResult{
		Browser: "Chrome",
		Domains: make(map[string]int),
	}

	script := `
tell application "Google Chrome"
	if it is running then
		set tabList to {}
		repeat with w in windows
			repeat with t in tabs of w
				set end of tabList to (title of t) & "|||" & (URL of t)
			end repeat
		end repeat
		set AppleScript's text item delimiters to ":::"
		set tabText to tabList as text
		set AppleScript's text item delimiters to ""
		return tabText
	end if
end tell
return ""
`

	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		result.Error = fmt.Errorf("chrome not running or unavailable: %w", err)
		return result
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return result
	}

	result.Available = true
	tabs := strings.Split(outputStr, ":::")

	for _, tab := range tabs {
		if tab == "" {
			continue
		}

		parts := strings.Split(tab, "|||")
		if len(parts) != 2 {
			continue
		}

		title := strings.TrimSpace(parts[0])
		urlStr := strings.TrimSpace(parts[1])

		domain := extractDomain(urlStr)

		result.Tabs = append(result.Tabs, BrowserTab{
			Title:  title,
			URL:    urlStr,
			Domain: domain,
		})
		result.TabCount++

		if domain != "" {
			result.Domains[domain]++
		}
	}

	return result
}

func collectSafariTabs(ctx context.Context) BrowserResult {
	result := BrowserResult{
		Browser: "Safari",
		Domains: make(map[string]int),
	}

	script := `
tell application "Safari"
	if it is running then
		set tabList to {}
		repeat with w in windows
			repeat with t in tabs of w
				set end of tabList to (name of t) & "|||" & (URL of t)
			end repeat
		end repeat
		set AppleScript's text item delimiters to ":::"
		set tabText to tabList as text
		set AppleScript's text item delimiters to ""
		return tabText
	end if
end tell
return ""
`

	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		result.Error = fmt.Errorf("safari not running or unavailable: %w", err)
		return result
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return result
	}

	result.Available = true
	tabs := strings.Split(outputStr, ":::")

	for _, tab := range tabs {
		if tab == "" {
			continue
		}

		parts := strings.Split(tab, "|||")
		if len(parts) != 2 {
			continue
		}

		title := strings.TrimSpace(parts[0])
		urlStr := strings.TrimSpace(parts[1])

		domain := extractDomain(urlStr)

		result.Tabs = append(result.Tabs, BrowserTab{
			Title:  title,
			URL:    urlStr,
			Domain: domain,
		})
		result.TabCount++

		if domain != "" {
			result.Domains[domain]++
		}
	}

	return result
}

func collectEdgeTabs(ctx context.Context) BrowserResult {
	result := BrowserResult{
		Browser: "Edge",
		Domains: make(map[string]int),
	}

	script := `
tell application "Microsoft Edge"
	if it is running then
		set tabList to {}
		repeat with w in windows
			repeat with t in tabs of w
				set end of tabList to (title of t) & "|||" & (URL of t)
			end repeat
		end repeat
		set AppleScript's text item delimiters to ":::"
		set tabText to tabList as text
		set AppleScript's text item delimiters to ""
		return tabText
	end if
end tell
return ""
`

	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		result.Error = fmt.Errorf("edge not running or unavailable: %w", err)
		return result
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return result
	}

	result.Available = true
	tabs := strings.Split(outputStr, ":::")

	for _, tab := range tabs {
		if tab == "" {
			continue
		}

		parts := strings.Split(tab, "|||")
		if len(parts) != 2 {
			continue
		}

		title := strings.TrimSpace(parts[0])
		urlStr := strings.TrimSpace(parts[1])

		domain := extractDomain(urlStr)

		result.Tabs = append(result.Tabs, BrowserTab{
			Title:  title,
			URL:    urlStr,
			Domain: domain,
		})
		result.TabCount++

		if domain != "" {
			result.Domains[domain]++
		}
	}

	return result
}

func extractDomain(urlStr string) string {
	if urlStr == "" {
		return ""
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	host := parsed.Host
	if host == "" {
		return ""
	}

	// Remove www. prefix
	host = strings.TrimPrefix(host, "www.")

	return host
}
