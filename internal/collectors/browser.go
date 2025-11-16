package collectors

import (
	"context"
	"database/sql"
	"fmt"
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

// IssueVisit represents a single issue/ticket visit
type IssueVisit struct {
	ID        string // e.g., "PROJ-123", "github.com/org/repo/issues/456"
	Tracker   string // e.g., "Jira", "GitHub", "Linear"
	URL       string // Full URL
	VisitCount int
}

// IssuesResult contains issue/ticket tracking information
type IssuesResult struct {
	Issues    []IssueVisit
	Available bool
	Error     error
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

// issuePattern represents a pattern for matching issue tracker URLs
type issuePattern struct {
	tracker string
	pattern *regexp.Regexp
	idGroup int // which capture group contains the ID
}

// Issue tracker URL patterns
var issuePatterns = []issuePattern{
	{
		tracker: "GitHub",
		pattern: regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/issues/(\d+)`),
		idGroup: 0, // Use full match as ID
	},
	{
		tracker: "Jira",
		pattern: regexp.MustCompile(`([^/]+\.)?atlassian\.net/browse/([A-Z]+-\d+)`),
		idGroup: 2, // Project key + number
	},
	{
		tracker: "Linear",
		pattern: regexp.MustCompile(`linear\.app/([^/]+/)?issue/([A-Z]+-[A-Z0-9]+)`),
		idGroup: 2, // Issue ID
	},
	{
		tracker: "GitLab",
		pattern: regexp.MustCompile(`gitlab\.com/([^/]+)/([^/]+)/-/issues/(\d+)`),
		idGroup: 0, // Use full match as ID
	},
	{
		tracker: "Azure DevOps",
		pattern: regexp.MustCompile(`dev\.azure\.com/([^/]+)/([^/]+)/_workitems/edit/(\d+)`),
		idGroup: 3, // Work item ID
	},
}

// CollectIssues collects issue/ticket URLs from browser history
func CollectIssues(ctx context.Context) IssuesResult {
	result := IssuesResult{}

	// Collect from Chrome, Safari, and Edge history
	issueMap := make(map[string]*IssueVisit)

	// Get today's start time (midnight)
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Chrome history
	chromeIssues := collectChromeHistory(ctx, todayStart)
	for _, issue := range chromeIssues {
		key := issue.Tracker + ":" + issue.ID
		if existing, ok := issueMap[key]; ok {
			existing.VisitCount += issue.VisitCount
		} else {
			issueMap[key] = &IssueVisit{
				ID:         issue.ID,
				Tracker:    issue.Tracker,
				URL:        issue.URL,
				VisitCount: issue.VisitCount,
			}
		}
	}

	// Safari history
	safariIssues := collectSafariHistory(ctx, todayStart)
	for _, issue := range safariIssues {
		key := issue.Tracker + ":" + issue.ID
		if existing, ok := issueMap[key]; ok {
			existing.VisitCount += issue.VisitCount
		} else {
			issueMap[key] = &IssueVisit{
				ID:         issue.ID,
				Tracker:    issue.Tracker,
				URL:        issue.URL,
				VisitCount: issue.VisitCount,
			}
		}
	}

	// Edge history
	edgeIssues := collectEdgeHistory(ctx, todayStart)
	for _, issue := range edgeIssues {
		key := issue.Tracker + ":" + issue.ID
		if existing, ok := issueMap[key]; ok {
			existing.VisitCount += issue.VisitCount
		} else {
			issueMap[key] = &IssueVisit{
				ID:         issue.ID,
				Tracker:    issue.Tracker,
				URL:        issue.URL,
				VisitCount: issue.VisitCount,
			}
		}
	}

	// Convert map to slice
	for _, issue := range issueMap {
		result.Issues = append(result.Issues, *issue)
	}

	// Sort by visit count (descending)
	for i := 0; i < len(result.Issues); i++ {
		for j := i + 1; j < len(result.Issues); j++ {
			if result.Issues[j].VisitCount > result.Issues[i].VisitCount {
				result.Issues[i], result.Issues[j] = result.Issues[j], result.Issues[i]
			}
		}
	}

	result.Available = len(result.Issues) > 0

	return result
}

// collectChromeHistory reads Chrome history database
func collectChromeHistory(ctx context.Context, since time.Time) []IssueVisit {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	historyPath := filepath.Join(homeDir, "Library", "Application Support", "Google", "Chrome", "Default", "History")
	return parseHistoryDB(ctx, historyPath, since, "chrome")
}

// collectSafariHistory reads Safari history database
func collectSafariHistory(ctx context.Context, since time.Time) []IssueVisit {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	historyPath := filepath.Join(homeDir, "Library", "Safari", "History.db")
	return parseSafariHistoryDB(ctx, historyPath, since)
}

// collectEdgeHistory reads Edge history database
func collectEdgeHistory(ctx context.Context, since time.Time) []IssueVisit {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	historyPath := filepath.Join(homeDir, "Library", "Application Support", "Microsoft Edge", "Default", "History")
	return parseHistoryDB(ctx, historyPath, since, "edge")
}

// parseHistoryDB parses Chrome/Edge-style history databases
func parseHistoryDB(ctx context.Context, dbPath string, since time.Time, browserType string) []IssueVisit {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil
	}

	// Copy database to temp location to avoid locking issues
	tempDB := filepath.Join(os.TempDir(), fmt.Sprintf("rekap_%s_history_%d.db", browserType, time.Now().Unix()))
	defer os.Remove(tempDB)

	// Use cp command to copy the file
	cmd := exec.CommandContext(ctx, "cp", dbPath, tempDB)
	if err := cmd.Run(); err != nil {
		return nil
	}

	db, err := sql.Open("sqlite", tempDB)
	if err != nil {
		return nil
	}
	defer db.Close()

	// Chrome/Edge use microseconds since Unix epoch
	sinceChrome := since.Unix() * 1000000

	query := `
		SELECT url, visit_count 
		FROM urls 
		WHERE last_visit_time >= ?
		ORDER BY visit_count DESC
	`

	rows, err := db.QueryContext(ctx, query, sinceChrome)
	if err != nil {
		return nil
	}
	defer rows.Close()

	return extractIssuesFromRows(rows)
}

// parseSafariHistoryDB parses Safari history database
func parseSafariHistoryDB(ctx context.Context, dbPath string, since time.Time) []IssueVisit {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil
	}

	// Copy database to temp location
	tempDB := filepath.Join(os.TempDir(), fmt.Sprintf("rekap_safari_history_%d.db", time.Now().Unix()))
	defer os.Remove(tempDB)

	cmd := exec.CommandContext(ctx, "cp", dbPath, tempDB)
	if err := cmd.Run(); err != nil {
		return nil
	}

	db, err := sql.Open("sqlite", tempDB)
	if err != nil {
		return nil
	}
	defer db.Close()

	// Safari uses Core Data timestamp (seconds since 2001-01-01)
	referenceDate := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	sinceSafari := since.Sub(referenceDate).Seconds()

	query := `
		SELECT 
			history_items.url,
			COUNT(history_visits.id) as visit_count
		FROM history_items
		LEFT JOIN history_visits ON history_items.id = history_visits.history_item
		WHERE history_visits.visit_time >= ?
		GROUP BY history_items.url
		ORDER BY visit_count DESC
	`

	rows, err := db.QueryContext(ctx, query, sinceSafari)
	if err != nil {
		return nil
	}
	defer rows.Close()

	return extractIssuesFromRows(rows)
}

// extractIssuesFromRows extracts issue URLs from database rows
func extractIssuesFromRows(rows *sql.Rows) []IssueVisit {
	var issues []IssueVisit
	issueMap := make(map[string]*IssueVisit)

	for rows.Next() {
		var urlStr string
		var visitCount int

		if err := rows.Scan(&urlStr, &visitCount); err != nil {
			continue
		}

		// Try to match against issue patterns
		for _, pattern := range issuePatterns {
			matches := pattern.pattern.FindStringSubmatch(urlStr)
			if matches != nil {
				var issueID string
				if pattern.idGroup == 0 {
					// Use the matched portion as ID
					issueID = matches[0]
				} else {
					// Use specific capture group
					issueID = matches[pattern.idGroup]
				}

				key := pattern.tracker + ":" + issueID
				if existing, ok := issueMap[key]; ok {
					existing.VisitCount += visitCount
				} else {
					issueMap[key] = &IssueVisit{
						ID:         issueID,
						Tracker:    pattern.tracker,
						URL:        urlStr,
						VisitCount: visitCount,
					}
				}
				break // Only match first pattern
			}
		}
	}

	// Convert map to slice
	for _, issue := range issueMap {
		issues = append(issues, *issue)
	}

	return issues
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
