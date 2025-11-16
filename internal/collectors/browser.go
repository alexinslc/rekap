package collectors

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// BrowserTab represents a single browser tab
type BrowserTab struct {
	Title  string
	URL    string
	Domain string
}

// BrowserResult contains browser tab information and history
type BrowserResult struct {
	Tabs            []BrowserTab
	TabCount        int
	Domains         map[string]int // domain -> tab count
	Browser         string
	Available       bool
	Error           error
	// History data
	URLsVisited     int
	TopDomain       string
	TopDomainVisits int
	IssueURLs       []string // Jira, GitHub, Linear issue URLs
	HistoryDomains  map[string]int // domain -> visit count from history
}

// BrowsersResult aggregates all browser data
type BrowsersResult struct {
	Chrome           BrowserResult
	Safari           BrowserResult
	Edge             BrowserResult
	TotalTabs        int
	TopDomains       map[string]int // aggregated across all browsers
	Available        bool
	// History aggregation
	TotalURLsVisited int
	AllIssueURLs     []string
	TopHistoryDomain string
	TopDomainVisits  int
}

// CollectBrowserTabs retrieves open tabs from Chrome, Safari, and Edge
// and also parses browser history for today's activity
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

	// Aggregate tab data
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

	// Aggregate history data
	result.TotalURLsVisited = result.Chrome.URLsVisited + result.Safari.URLsVisited + result.Edge.URLsVisited
	
	// Combine all issue URLs, deduplicated
	issueURLSet := make(map[string]struct{})
	for _, url := range result.Chrome.IssueURLs {
		issueURLSet[url] = struct{}{}
	}
	for _, url := range result.Safari.IssueURLs {
		issueURLSet[url] = struct{}{}
	}
	for _, url := range result.Edge.IssueURLs {
		issueURLSet[url] = struct{}{}
	}
	result.AllIssueURLs = make([]string, 0, len(issueURLSet))
	for url := range issueURLSet {
		result.AllIssueURLs = append(result.AllIssueURLs, url)
	}
	// Find top history domain across all browsers
	allHistoryDomains := make(map[string]int)
	for domain, count := range result.Chrome.HistoryDomains {
		allHistoryDomains[domain] += count
	}
	for domain, count := range result.Safari.HistoryDomains {
		allHistoryDomains[domain] += count
	}
	for domain, count := range result.Edge.HistoryDomains {
		allHistoryDomains[domain] += count
	}
	
	// Find top domain
	maxVisits := 0
	for domain, count := range allHistoryDomains {
		if count > maxVisits {
			maxVisits = count
			result.TopHistoryDomain = domain
			result.TopDomainVisits = count
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
	result := collectBrowserTabsForApp(ctx, "Chrome", "Google Chrome", "title of t")
	
	// Also collect history
	historyData := collectChromeHistory(ctx)
	result.URLsVisited = historyData.URLsVisited
	result.TopDomain = historyData.TopDomain
	result.TopDomainVisits = historyData.TopDomainVisits
	result.IssueURLs = historyData.IssueURLs
	result.HistoryDomains = historyData.HistoryDomains
	
	return result
}

func collectSafariTabs(ctx context.Context) BrowserResult {
	result := collectBrowserTabsForApp(ctx, "Safari", "Safari", "name of t")
	
	// Also collect history
	historyData := collectSafariHistory(ctx)
	result.URLsVisited = historyData.URLsVisited
	result.TopDomain = historyData.TopDomain
	result.TopDomainVisits = historyData.TopDomainVisits
	result.IssueURLs = historyData.IssueURLs
	result.HistoryDomains = historyData.HistoryDomains
	
	return result
}

func collectEdgeTabs(ctx context.Context) BrowserResult {
	result := collectBrowserTabsForApp(ctx, "Edge", "Microsoft Edge", "title of t")
	
	// Also collect history
	historyData := collectEdgeHistory(ctx)
	result.URLsVisited = historyData.URLsVisited
	result.TopDomain = historyData.TopDomain
	result.TopDomainVisits = historyData.TopDomainVisits
	result.IssueURLs = historyData.IssueURLs
	result.HistoryDomains = historyData.HistoryDomains
	
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

// BrowserHistoryData contains history-specific data
type BrowserHistoryData struct {
	URLsVisited     int
	TopDomain       string
	TopDomainVisits int
	IssueURLs       []string
	HistoryDomains  map[string]int
}

// collectChromeHistory parses Chrome history database
func collectChromeHistory(ctx context.Context) BrowserHistoryData {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return BrowserHistoryData{}
	}

	historyPath := filepath.Join(homeDir, "Library", "Application Support", "Google", "Chrome", "Default", "History")
	return collectBrowserHistory(ctx, historyPath, "chrome")
}

// collectSafariHistory parses Safari history database
func collectSafariHistory(ctx context.Context) BrowserHistoryData {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return BrowserHistoryData{}
	}

	historyPath := filepath.Join(homeDir, "Library", "Safari", "History.db")
	return collectBrowserHistory(ctx, historyPath, "safari")
}

// collectEdgeHistory parses Edge history database
func collectEdgeHistory(ctx context.Context) BrowserHistoryData {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return BrowserHistoryData{}
	}

	historyPath := filepath.Join(homeDir, "Library", "Application Support", "Microsoft Edge", "Default", "History")
	return collectBrowserHistory(ctx, historyPath, "edge")
}

// collectBrowserHistory is a generic function to collect history from Chrome/Edge/Safari databases
func collectBrowserHistory(ctx context.Context, dbPath, browserType string) BrowserHistoryData {
	result := BrowserHistoryData{
		HistoryDomains: make(map[string]int),
	}

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return result
	}

	// Copy database to temp location to avoid lock issues
	tempDB, err := copyToTemp(dbPath)
	if err != nil {
		return result
	}
	defer os.Remove(tempDB)

	// Open the database
	db, err := sql.Open("sqlite", tempDB)
	if err != nil {
		return result
	}
	defer db.Close()

	// Get today's timestamp range
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var rows *sql.Rows
	if browserType == "safari" {
		// Safari uses Core Data timestamp (seconds since 2001-01-01)
		coreDataEpoch := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
		startTimestamp := midnight.Sub(coreDataEpoch).Seconds()
		endTimestamp := now.Sub(coreDataEpoch).Seconds()
		
		// Join history_items and history_visits to get all visits for today
		query := `
			SELECT hi.url, COUNT(*) as today_visit_count
			FROM history_items hi
			JOIN history_visits hv ON hi.id = hv.history_item
			WHERE hv.visit_time >= ? AND hv.visit_time < ?
			GROUP BY hi.url
			ORDER BY today_visit_count DESC
		`
		rows, err = db.QueryContext(ctx, query, startTimestamp, endTimestamp)
	} else {
		// Chrome/Edge use microseconds since Unix epoch
		startTimestamp := midnight.UnixMicro()
		
		query := `
			SELECT url, visit_count
			FROM urls
			WHERE last_visit_time >= ?
			ORDER BY visit_count DESC
		`
		rows, err = db.QueryContext(ctx, query, startTimestamp)
	}

	if err != nil {
		return result
	}
	defer rows.Close()

	// Process results
	for rows.Next() {
		var urlStr string
		var visitCount int

		if err := rows.Scan(&urlStr, &visitCount); err != nil {
			continue
		}

		result.URLsVisited++

		// Extract domain
		domain := extractDomain(urlStr)
		if domain != "" {
			result.HistoryDomains[domain] += visitCount
		}

		// Check if it's an issue URL
		if isIssueURL(urlStr) {
			result.IssueURLs = append(result.IssueURLs, urlStr)
		}
	}

	// Find top domain
	maxVisits := 0
	for domain, count := range result.HistoryDomains {
		if count > maxVisits {
			maxVisits = count
			result.TopDomain = domain
			result.TopDomainVisits = count
		}
	}

	return result
}

// copyToTemp copies a file to a temporary location
func copyToTemp(srcPath string) (string, error) {
	src, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer src.Close()

	tmpFile, err := os.CreateTemp("", "browser-history-*.db")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, src)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

// Precompiled regexes for common issue trackers
var issueURLRegexes = []*regexp.Regexp{
	// Jira: project-123, PROJ-456
	regexp.MustCompile(`[A-Z]+-\d+`),
	// GitHub: /issues/, /pull/
	regexp.MustCompile(`github\.com/.+/(issues|pull)/\d+`),
	// Linear: /issue/
	regexp.MustCompile(`linear\.app/.+/issue/`),
	// GitLab: /issues/, /merge_requests/
	regexp.MustCompile(`gitlab\.com/.+/(issues|merge_requests)/\d+`),
	// Bitbucket: /issues/
	regexp.MustCompile(`bitbucket\.org/.+/issues/\d+`),
	// Azure DevOps: /_workitems/ or /workitems/
	regexp.MustCompile(`dev\.azure\.com/.+/_?workitems/\d+`),
}

// isIssueURL checks if a URL is an issue/ticket URL
func isIssueURL(urlStr string) bool {
	for _, re := range issueURLRegexes {
		if re.MatchString(urlStr) {
			return true
		}
	}
	return false
}

// FormatIssueURLs formats a list of issue URLs for display
func FormatIssueURLs(issueURLs []string) string {
	if len(issueURLs) == 0 {
		return ""
	}
	
	// Show up to 3 issues
	limit := 3
	if len(issueURLs) > limit {
		return strings.Join(issueURLs[:limit], ", ") + fmt.Sprintf(" (+%d more)", len(issueURLs)-limit)
	}
	
	return strings.Join(issueURLs, ", ")
}
