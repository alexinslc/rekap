package collectors

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// MediaResult contains now playing information
type MediaResult struct {
	Track           string
	App             string
	DurationMinutes int
	Available       bool
	Error           error
}

// CollectMedia retrieves currently or last played media information
func CollectMedia(ctx context.Context) MediaResult {
	result := MediaResult{Available: false}

	// Try using osascript to query Music app
	cmd := exec.CommandContext(ctx, "osascript", "-e", `
		tell application "Music"
			if it is running then
				if player state is not stopped then
					set trackName to name of current track
					set appName to "Music"
					return trackName & "|" & appName
				end if
			end if
		end tell
		return ""
	`)
	
	output, err := cmd.Output()
	if err == nil {
		outputStr := strings.TrimSpace(string(output))
		if outputStr != "" {
			parts := strings.Split(outputStr, "|")
			if len(parts) >= 2 {
				result.Track = parts[0]
				result.App = parts[1]
				result.Available = true
				// Duration tracking would require persistent monitoring
				result.DurationMinutes = 0
				return result
			}
		}
	}

	// Try Spotify via osascript
	cmd = exec.CommandContext(ctx, "osascript", "-e", `
		tell application "Spotify"
			if it is running then
				if player state is playing then
					set trackName to name of current track
					set artistName to artist of current track
					return trackName & " - " & artistName & "|Spotify"
				end if
			end if
		end tell
		return ""
	`)
	
	output, err = cmd.Output()
	if err == nil {
		outputStr := strings.TrimSpace(string(output))
		if outputStr != "" {
			parts := strings.Split(outputStr, "|")
			if len(parts) >= 2 {
				result.Track = parts[0]
				result.App = parts[1]
				result.Available = true
				result.DurationMinutes = 0
				return result
			}
		}
	}

	// Check if nowplaying-cli is available
	cmd = exec.CommandContext(ctx, "nowplaying-cli", "get", "title")
	titleOutput, titleErr := cmd.Output()
	
	cmd = exec.CommandContext(ctx, "nowplaying-cli", "get", "artist")
	artistOutput, artistErr := cmd.Output()
	
	cmd = exec.CommandContext(ctx, "nowplaying-cli", "get", "app")
	appOutput, appErr := cmd.Output()
	
	if titleErr == nil && appErr == nil {
		title := strings.TrimSpace(string(titleOutput))
		app := strings.TrimSpace(string(appOutput))
		
		if title != "" && app != "" {
			track := title
			if artistErr == nil {
				artist := strings.TrimSpace(string(artistOutput))
				if artist != "" {
					track = fmt.Sprintf("%s - %s", title, artist)
				}
			}
			
			result.Track = track
			result.App = app
			result.Available = true
			result.DurationMinutes = 0
			return result
		}
	}

	result.Error = fmt.Errorf("no media playing or media info unavailable")
	return result
}
