// Package main implements a JSON file monitor that watches for new entries
// and displays them in real-time, similar to the 'tail' command but for JSON files.
// This is a very basic implementation and only works with JSON files containing
// an array of strings. It also expects that the array of strings is updated sequentially
// with new entries at the end of the file. If the file is modified in other ways,
// the behavior may be unpredictable.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/siliconchaos/json-tail/spinner"
)

// Config holds the configuration for the application.
// In Go, it's common to group related configuration fields in a struct.
type Config struct {
	filename string  // The JSON file path to monitor
	interval float64 // The time interval (in seconds) between file checks
}

func main() {
	// Configure logging to include timestamps
	// Log flags are bit flags combined using the OR operator (|)
	log.SetFlags(log.Ldate | log.Ltime)

	// Parse and validate command-line flags
	config := parseFlags()

	// Ensure the target file exists and is accessible
	if err := validateFile(config.filename); err != nil {
		log.Fatal(err)
	}

	// Convert to absolute path to handle relative path inputs
	// This ensures consistent file access regardless of working directory
	absPath, err := filepath.Abs(config.filename)
	if err != nil {
		log.Fatal("Error getting absolute path: ", err)
	}

	// Set up graceful shutdown handling
	// Buffer size of 1 ensures we don't miss the interrupt signal
	// Mainly used to handle Ctrl+C (os.Interrupt) but can be extended
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Channel for communicating new entries between goroutines
	// Unbuffered channel ensures synchronous communication
	changes := make(chan []string)

	// Create and start the spnr
	spnr := spinner.New([]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"})
	// spinner.SetState("Waiting for changes...")
	// spinner.Start()

	// Display initial file contents
	var initialLength int
	if entries, err := readJSONFile(absPath); err == nil {
		// Temporarily stop the spinner for initial output
		// spinner.Stop()

		fmt.Println("Initial entries (last 10)")
		fmt.Println("----------------------------")
		for _, entry := range lastN(entries, 10) {
			fmt.Println(entry)
		}
		fmt.Println("----------------------------")
		fmt.Printf("Monitoring file for new entries (checking every %.1f seconds)...\n\n", config.interval)
		initialLength = len(entries)

		// start the spinner
		spnr.SetState("Waiting for changes...")
		spnr.Start()
	}

	// Start the file monitor in a separate goroutine
	// This allows the main goroutine to handle user interrupts
	go monitorFile(absPath, config.interval, changes, initialLength, spnr)

	// Main event loop using select for concurrent channel operations
	// Select blocks until one of its cases can proceed
	for {
		select {
		case newEntries := <-changes:
			// Temporarily stop the spinner to display new entries
			spnr.Stop()
			// Process and display new entries as they arrive
			for _, entry := range newEntries {
				fmt.Printf("[%s] %s\n", time.Now().Format(time.RFC3339), entry)
			}
			// Restart the spinner after displaying new entries
			spnr.SetState("Waiting for changes...")
			spnr.Start()
		case <-sigChan:
			// Handle graceful shutdown on interrupt (Ctrl+C)
			spnr.Stop()
			fmt.Println("\nReceived interrupt signal, exiting...")
			return
		}
	}
}

// parseFlags processes command-line arguments and returns a Config struct.
// It handles both short (-i) and long (--interval) flag formats.
func parseFlags() Config {
	var config Config

	// Define command-line flags
	// The flag package automatically generates help text (-h or --help)
	const intervalHelp = "the interval in seconds at which to check the file for changes"
	flag.Float64Var(&config.interval, "i", 1.0, intervalHelp)
	flag.Float64Var(&config.interval, "interval", 1.0, intervalHelp)

	// Parse command-line flags
	flag.Parse()

	// Get non-flag arguments (expected to be the filename)
	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("Please provide a JSON file to monitor\nUsage: json-tail <filename> [-i <interval>]")
	}

	config.filename = args[0]

	// Validate interval value
	if config.interval <= 0 {
		log.Fatal("Interval must be a positive number")
	}

	return config
}

// validateFile checks if the specified file exists, is readable,
// and is not a directory.
func validateFile(filename string) error {
	info, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", filename)
		}
		return fmt.Errorf("error accessing file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("%s is a directory", filename)
	}

	// Verify read permissions by attempting to open the file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close() // Ensure file is closed after function returns

	return nil
}

// monitorFile watches the specified file for changes and sends new entries
// through the changes channel. It runs continuously until the program exits.
// The changes parameter is a send-only channel (chan<-) as this function
// only sends data and never receives from the channel.
func monitorFile(filename string, interval float64, changes chan<- []string, previousLength int, spinner *spinner.Spinner) {
	// Create a ticker for regular interval checks
	// time.Duration is a type representing nanosecond precision
	ticker := time.NewTicker(time.Duration(interval * float64(time.Second)))
	defer ticker.Stop() // Ensure ticker is stopped when function returns

	// Loop indefinitely, checking for new entries on each tick
	for range ticker.C {
		spinner.SetState("Checking for changes...")
		entries, err := readJSONFile(filename)
		if err != nil {
			log.Printf("Error reading file: %v", err)
			spinner.SetState("Error reading file")
			continue
		}

		// If new entries are found, send them through the channel
		if len(entries) > previousLength {
			newEntries := entries[previousLength:]
			changes <- newEntries
			previousLength = len(entries)
		}
		spinner.SetState("Waiting for changes...")
	}
}

// lastN returns the last n elements of a slice.
// If the slice has fewer than n elements, it returns the entire slice.
func lastN(entries []string, n int) []string {
	if len(entries) <= n {
		return entries
	}
	return entries[len(entries)-n:]
}

// readJSONFile reads and parses a JSON file containing an array of strings.
// It returns the parsed entries and any error encountered during reading
// or parsing.
func readJSONFile(filename string) ([]string, error) {
	// Read entire file into memory
	// TODO: For large files, you might want to use streaming JSON decoding instead
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var entries []string
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return entries, nil
}
