// Package spinner provides a simple terminal spinner for showing progress
// and status updates in command-line applications. It supports custom animation
// frames and concurrent status message updates.
package spinner

import (
	"fmt"
	"sync"
	"time"
)

// Spinner represents a terminal spinner animation with status message support.
// It is safe for concurrent use.
type Spinner struct {
	frames []string   // Animation frames to cycle through
	mu     sync.Mutex // Mutex for thread-safe updates
	active bool       // Current state of the spinner
	state  string     // Current state message
}

// New creates a new spinner with the given animation frames.
// If no frames are provided, it will use the default frames.
func New(frames []string) *Spinner {
	if len(frames) == 0 {
		frames = DefaultFrames()
	}
	return &Spinner{
		frames: frames,
		active: true,
	}
}

// SetState updates the current state message in a thread-safe manner.
// This can be called while the spinner is active to update its message.
func (s *Spinner) SetState(state string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = state
}

// Stop stops the spinner animation and clears the current line.
// It is safe to call Stop multiple times.
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Clear the current line and set the spinner to inactive
	fmt.Print("\r\033[K")
	s.active = false
}

// Start begins the spinner animation in a separate goroutine.
// The spinner will continue until Stop is called.
// It is safe to call Start after Stop to restart the animation.
func (s *Spinner) Start() {
	s.mu.Lock()
	s.active = true
	s.mu.Unlock()

	// Clear the current line and move cursor to beginning
	fmt.Print("\r\033[K")

	go func() {
		for i := 0; s.active; i++ {
			s.mu.Lock()
			// Print the spinner frame and state message
			frame := s.frames[i%len(s.frames)]
			fmt.Printf("\r%s %s", frame, s.state)
			s.mu.Unlock()

			time.Sleep(100 * time.Millisecond)
		}
		// Clear the spinner line when done
		fmt.Print("\r\033[K")
	}()
}

// DefaultFrames returns a set of default braille-based spinner frames
// that provide a smooth animation in terminal output.
func DefaultFrames() []string {
	return []string{
		"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
	}
}

// Example usage:
/*
   s := spinner.New(spinner.DefaultFrames())
   s.SetState("Processing...")
   s.Start()

   // Do some work
   time.Sleep(2 * time.Second)

   s.SetState("Almost done...")
   time.Sleep(1 * time.Second)

   s.Stop()
*/
