package collectors

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"strings"

	"github.com/alexinslc/rekap/internal/config"
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
	Chrome            BrowserResult
	Safari            BrowserResult
	Edge              BrowserResult
	TotalTabs         int
	TopDomains        map[string]int // aggregated across all browsers
	WorkVisits        int
	DistractionVisits int
	NeutralVisits     int
	Available         bool
}

// CollectBrowserTabs retrieves open tabs from Chrome, Safari, and Edge
func CollectBrowserTabs(ctx context.Context, cfg *config.Config) BrowsersResult {
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

	// Categorize domains if config is provided
	if cfg != nil {
		for domain, count := range result.TopDomains {
			category := cfg.CategorizeDomain(domain)
			switch category {
			case "work":
				result.WorkVisits += count
			case "distraction":
				result.DistractionVisits += count
			case "neutral":
				result.NeutralVisits += count
			default:
				result.NeutralVisits += count
			}
		}
	}

	result.Available = result.Chrome.Available || result.Safari.Available || result.Edge.Available

	return result
}

// collectBrowserTabsForApp is a generic helper to collect browser tabs
// browserName: display name for the browser (e.g., "Chrome")
// appName: AppleScript application name (e.g., "Google Chrome")
// titleProperty: AppleScript property for tab title ("title of t" or "name of t")
func collectBrowserTabsForApp(ctx context.Context, browserName, appName, titleProperty string) BrowserResult {
	result := BrowserResult{
		Browser: browserName,
		Domains: make(map[string]int),
	}

	script := fmt.Sprintf(`
tell application "%s"
	if it is running then
		set tabList to {}
		repeat with w in windows
			repeat with t in tabs of w
				set end of tabList to (%s) & "|||" & (URL of t)
			end repeat
		end repeat
		set AppleScript's text item delimiters to ":::"
		set tabText to tabList as text
		set AppleScript's text item delimiters to ""
		return tabText
	end if
end tell
return ""
`, appName, titleProperty)

	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		result.Error = fmt.Errorf("%s not running or unavailable: %w", strings.ToLower(browserName), err)
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

func collectChromeTabs(ctx context.Context) BrowserResult {
	return collectBrowserTabsForApp(ctx, "Chrome", "Google Chrome", "title of t")
}

func collectSafariTabs(ctx context.Context) BrowserResult {
	return collectBrowserTabsForApp(ctx, "Safari", "Safari", "name of t")
}

func collectEdgeTabs(ctx context.Context) BrowserResult {
	return collectBrowserTabsForApp(ctx, "Edge", "Microsoft Edge", "title of t")
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
