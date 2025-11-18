package ui

import (
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

var spinnerStyle = lipgloss.NewStyle().Foreground(primaryColor)

// LoadingSpinner wraps the bubbles spinner for simple loading states
type LoadingSpinner struct {
	spinner  spinner.Model
	message  string
	done     bool
	mu       sync.Mutex
	ticker   *time.Ticker
	stopChan chan bool
}

// NewLoadingSpinner creates a new loading spinner with the given message
func NewLoadingSpinner(message string) *LoadingSpinner {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return &LoadingSpinner{
		spinner:  s,
		message:  message,
		stopChan: make(chan bool),
	}
}

// Start begins the spinner animation in a goroutine
func (ls *LoadingSpinner) Start() {
	if !IsTTY() {
		return
	}

	ls.ticker = time.NewTicker(100 * time.Millisecond)

	go func() {
		for {
			select {
			case <-ls.stopChan:
				ls.ticker.Stop()
				return
			case <-ls.ticker.C:
				ls.mu.Lock()
				if !ls.done {
					fmt.Printf("\r%s %s  ", ls.spinner.View(), ls.message)
					ls.spinner, _ = ls.spinner.Update(ls.spinner.Tick())
				}
				ls.mu.Unlock()
			}
		}
	}()
}

// UpdateMessage updates the spinner message
func (ls *LoadingSpinner) UpdateMessage(message string) {
	if !IsTTY() {
		return
	}

	ls.mu.Lock()
	ls.message = message
	ls.mu.Unlock()
}

// Stop stops the spinner and clears the line
func (ls *LoadingSpinner) Stop() {
	if !IsTTY() {
		return
	}

	ls.mu.Lock()
	ls.done = true
	ls.mu.Unlock()

	// Send stop signal and wait briefly for goroutine to finish
	close(ls.stopChan)
	time.Sleep(50 * time.Millisecond)

	// Clear the line
	fmt.Print("\r\033[K")
}

// ShowLoadingWithSpinner is a convenience function that creates and starts a spinner
func ShowLoadingWithSpinner(message string) *LoadingSpinner {
	spinner := NewLoadingSpinner(message)
	spinner.Start()
	return spinner
}
