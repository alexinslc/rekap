package ui

import (
	"fmt"
	"sync"
	"time"
)

// Spinner represents a simple loading spinner
type Spinner struct {
	frames   []string
	current  int
	mu       sync.Mutex
	done     bool
	interval time.Duration
}

// NewSpinner creates a new spinner
func NewSpinner() *Spinner {
	return &Spinner{
		frames:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		interval: 80 * time.Millisecond,
	}
}

// Start begins the spinner animation
func (s *Spinner) Start(label string) {
	if !IsTTY() {
		fmt.Printf("%s...\n", label)
		return
	}

	go func() {
		for {
			s.mu.Lock()
			if s.done {
				s.mu.Unlock()
				return
			}
			frame := s.frames[s.current]
			s.current = (s.current + 1) % len(s.frames)
			s.mu.Unlock()

			fmt.Printf("\r%s %s", frame, label)
			time.Sleep(s.interval)
		}
	}()
}

// Stop stops the spinner
func (s *Spinner) Stop(success bool, finalLabel string) {
	s.mu.Lock()
	s.done = true
	s.mu.Unlock()

	if !IsTTY() {
		return
	}

	symbol := "✓"
	if !success {
		symbol = "✗"
	}

	fmt.Printf("\r%s %s\n", symbol, finalLabel)
}
