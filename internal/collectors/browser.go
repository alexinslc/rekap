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
	"sort"
	"strings"
	"time"

	"github.com/alexinslc/rekap/internal/config"
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
	Tabs      []BrowserTab
	TabCount  int
	Domains   map[string]int // domain -> tab count
	Browser   string
	Available bool
	Error     error
	// History data
	URLsVisited     int
	TopDomain       string
	TopDomainVisits int
	IssueURLs       []string       // Jira, GitHub, Linear issue URLs
	HistoryDomains  map[string]int // domain -> visit count from history
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
	// History aggregation
	TotalURLsVisited int
	AllIssueURLs     []string
	TopHistoryDomain string
	TopDomainVisits  int
}

// IssueVisit represents a single issue/ticket visit
type IssueVisit struct {
	ID         string // e.g., "PROJ-123", "github.com/org/repo/issues/456"
	Tracker    string // e.g., "Jira", "GitHub", "Linear"
	URL        string // Full URL
	VisitCount int
}

// IssuesResult contains issue/ticket tracking information
type IssuesResult struct {
	Issues    []IssueVisit
	Available bool
	Error     error
}

// CollectBrowserTabs retrieves open tabs from Chrome, Safari, and Edge
// and also parses browser history for today's activity
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

// mergeIssues merges issues into the issue map, aggregating visit counts
func mergeIssues(issueMap map[string]*IssueVisit, issues []IssueVisit) {
	for _, issue := range issues {
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
}

// CollectIssues collects issue/ticket URLs from browser history
func CollectIssues(ctx context.Context) IssuesResult {
	result := IssuesResult{}

	// Collect from Chrome, Safari, and Edge history
	issueMap := make(map[string]*IssueVisit)

	// Get today's start time (midnight)
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Merge issues from all browsers
	mergeIssues(issueMap, collectChromeIssues(ctx, todayStart))
	mergeIssues(issueMap, collectSafariIssues(ctx, todayStart))
	mergeIssues(issueMap, collectEdgeIssues(ctx, todayStart))

	// Convert map to slice
	for _, issue := range issueMap {
		result.Issues = append(result.Issues, *issue)
	}

	// Sort by visit count (descending)
	sort.Slice(result.Issues, func(i, j int) bool {
		return result.Issues[i].VisitCount > result.Issues[j].VisitCount
	})

	result.Available = len(result.Issues) > 0

	return result
}

// collectChromeIssues reads Chrome history database for issue URLs
func collectChromeIssues(ctx context.Context, since time.Time) []IssueVisit {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	historyPath := filepath.Join(homeDir, "Library", "Application Support", "Google", "Chrome", "Default", "History")
	return parseHistoryDB(ctx, historyPath, since, "chrome")
}

// collectSafariIssues reads Safari history database for issue URLs
func collectSafariIssues(ctx context.Context, since time.Time) []IssueVisit {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	historyPath := filepath.Join(homeDir, "Library", "Safari", "History.db")
	return parseSafariHistoryDB(ctx, historyPath, since)
}

// collectEdgeIssues reads Edge history database for issue URLs
func collectEdgeIssues(ctx context.Context, since time.Time) []IssueVisit {
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
	tempFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("rekap_%s_history_*.db", browserType))
	if err != nil {
		return nil
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close() // Close before copying to it

	// Use cp command to copy the file
	cmd := exec.CommandContext(ctx, "cp", dbPath, tempFile.Name())
	if err := cmd.Run(); err != nil {
		return nil
	}

	db, err := sql.Open("sqlite", tempFile.Name())
	if err != nil {
		return nil
	}
	defer db.Close()

	// Chrome/Edge use microseconds since January 1, 1601 (Windows epoch)
	windowsEpoch := time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)
	sinceChrome := since.Sub(windowsEpoch).Microseconds()

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
	tempFile, err := os.CreateTemp(os.TempDir(), "rekap_safari_history_*.db")
	if err != nil {
		return nil
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close() // Close before copying to it

	cmd := exec.CommandContext(ctx, "cp", dbPath, tempFile.Name())
	if err := cmd.Run(); err != nil {
		return nil
	}

	db, err := sql.Open("sqlite", tempFile.Name())
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

	// Convert map to slice, filtering out zero-visit entries
	for _, issue := range issueMap {
		if issue.VisitCount > 0 {
			issues = append(issues, *issue)
		}
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
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

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
		endTimestamp := now.UnixMicro()

		// Query visits table joined with urls for accurate today-only tracking
		query := `
			SELECT u.url, COUNT(*) as today_visit_count
			FROM urls u
			JOIN visits v ON u.id = v.url
			WHERE v.visit_time >= ? AND v.visit_time < ?
			GROUP BY u.url
			ORDER BY today_visit_count DESC
		`
		rows, err = db.QueryContext(ctx, query, startTimestamp, endTimestamp)
	}

	if err != nil {
		return result
	}
	defer rows.Close()

	// Process results - use map to deduplicate issue IDs
	issueIDSet := make(map[string]struct{})

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

		// Check if it's an issue URL and deduplicate
		if isIssueURL(urlStr) {
			issueID := extractIssueIdentifier(urlStr)
			issueIDSet[issueID] = struct{}{}
		}
	}

	// Convert deduplicated issue IDs to slice
	result.IssueURLs = make([]string, 0, len(issueIDSet))
	for issueID := range issueIDSet {
		result.IssueURLs = append(result.IssueURLs, issueID)
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
	// Jira: https://jira.example.com/browse/PROJ-123
	regexp.MustCompile(`(atlassian\.net|jira\.[^/]+)/browse/[A-Z]+-\d+`),
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

// extractIssueIdentifier extracts a clean issue identifier from a URL
func extractIssueIdentifier(urlStr string) string {
	// Jira: extract PROJ-123 from URL
	jiraRe := regexp.MustCompile(`/browse/([A-Z]+-\d+)`)
	if matches := jiraRe.FindStringSubmatch(urlStr); len(matches) > 1 {
		return matches[1]
	}

	// GitHub: extract owner/repo#123 from issues or pulls
	githubIssueRe := regexp.MustCompile(`github\.com/([^/]+/[^/]+)/(issues|pull)/(\d+)`)
	if matches := githubIssueRe.FindStringSubmatch(urlStr); len(matches) > 3 {
		return matches[1] + "#" + matches[3]
	}

	// Linear: extract issue ID from URL
	linearRe := regexp.MustCompile(`linear\.app/[^/]+/issue/([^/\?]+)`)
	if matches := linearRe.FindStringSubmatch(urlStr); len(matches) > 1 {
		return matches[1]
	}

	// GitLab: extract owner/repo#123
	gitlabRe := regexp.MustCompile(`gitlab\.com/([^/]+/[^/]+)/(issues|merge_requests)/(\d+)`)
	if matches := gitlabRe.FindStringSubmatch(urlStr); len(matches) > 3 {
		issueType := "!"
		if matches[2] == "issues" {
			issueType = "#"
		}
		return matches[1] + issueType + matches[3]
	}

	// Bitbucket: extract owner/repo#123
	bitbucketRe := regexp.MustCompile(`bitbucket\.org/([^/]+/[^/]+)/issues/(\d+)`)
	if matches := bitbucketRe.FindStringSubmatch(urlStr); len(matches) > 2 {
		return matches[1] + "#" + matches[2]
	}

	// Azure DevOps: extract workitem ID
	azureRe := regexp.MustCompile(`dev\.azure\.com/[^/]+/[^/]+/_?workitems/(\d+)`)
	if matches := azureRe.FindStringSubmatch(urlStr); len(matches) > 1 {
		return "WI-" + matches[1]
	}

	// Fallback: return the URL as-is
	return urlStr
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
