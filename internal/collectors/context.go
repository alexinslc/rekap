package collectors

import "fmt"

// ContextOverload represents detected context overload conditions
type ContextOverload struct {
	IsOverloaded   bool
	ActiveApps     int
	TotalTabs      int
	UniqueDomains  int
	WarningMessage string
}

// CheckContextOverload detects if the user has too many contexts active
// Triggers:
// - >5 apps actively used (not just open)
// - >30 browser tabs open
// - >10 unique domains from open tabs
func CheckContextOverload(apps AppsResult, browsers BrowsersResult) ContextOverload {
	result := ContextOverload{}

	// Count actively used apps (apps with recorded usage)
	activeApps := len(apps.TopApps)

	// Count total browser tabs
	totalTabs := browsers.TotalTabs

	// Count unique domains
	uniqueDomains := len(browsers.TopDomains)

	// Check thresholds
	appsOverload := activeApps > 5
	tabsOverload := totalTabs > 30
	domainsOverload := uniqueDomains > 10

	// If any threshold is exceeded, flag as overloaded
	if appsOverload || tabsOverload || domainsOverload {
		result.IsOverloaded = true
		result.ActiveApps = activeApps
		result.TotalTabs = totalTabs
		result.UniqueDomains = uniqueDomains

		// Build warning message
		switch {
		case appsOverload && tabsOverload:
			result.WarningMessage = formatWithCount(activeApps, "app") + " + " + formatWithCount(totalTabs, "tab") + " active"
		case appsOverload:
			result.WarningMessage = formatWithCount(activeApps, "app") + " active"
		case tabsOverload:
			result.WarningMessage = formatWithCount(totalTabs, "tab") + " active"
		case domainsOverload:
			result.WarningMessage = formatWithCount(uniqueDomains, "domain") + " active"
		}
	}

	return result
}

func formatWithCount(count int, singular string) string {
	if count == 1 {
		return "1 " + singular
	}
	return fmt.Sprintf("%d", count) + " " + singular + "s"
}
