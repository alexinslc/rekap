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
		if appsOverload && tabsOverload {
			result.WarningMessage = formatOverloadMessage(activeApps, totalTabs)
		} else if appsOverload {
			result.WarningMessage = formatOverloadMessage(activeApps, 0)
		} else if tabsOverload {
			result.WarningMessage = formatOverloadMessage(0, totalTabs)
		} else if domainsOverload {
			result.WarningMessage = formatDomainsOverloadMessage(uniqueDomains)
		}
	}

	return result
}

func formatOverloadMessage(activeApps, totalTabs int) string {
	if activeApps > 0 && totalTabs > 0 {
		return formatAppsTabs(activeApps, totalTabs)
	} else if activeApps > 0 {
		return formatAppsOnly(activeApps)
	} else if totalTabs > 0 {
		return formatTabsOnly(totalTabs)
	}
	return ""
}

func formatAppsTabs(apps, tabs int) string {
	return formatWithCount(apps, "app") + " + " + formatWithCount(tabs, "tab") + " active"
}

func formatAppsOnly(apps int) string {
	return formatWithCount(apps, "app") + " active"
}

func formatTabsOnly(tabs int) string {
	return formatWithCount(tabs, "tab") + " active"
}

func formatWithCount(count int, singular string) string {
	if count == 1 {
		return "1 " + singular
	}
	return fmt.Sprintf("%d", count) + " " + singular + "s"
}

func formatDomainsOverloadMessage(domains int) string {
	return formatWithCount(domains, "domain") + " active"
}
