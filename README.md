# JSON-Tail

⚠️ **Disclaimer: This is a learning exercise!** ⚠️

This project is a personal exercise to learn Go programming. It's a toy application created to understand Go's features like goroutines, channels, and file operations. It should not be used in production environments or for any critical tasks.

## What it does

Monitors a JSON file for new entries in real-time, similar to the `tail` command but specifically for JSON array files. The program expects a very specific format (an array of strings) and makes several assumptions about how the file is updated. It may break in unexpected ways with different JSON structures or update patterns.

## Features

- Basic real-time monitoring of JSON files
- Displays new entries as they are added
- Configurable check interval
- Simple spinner animation showing monitoring status
- Handles Ctrl+C gracefully

## Limitations

- Only works with JSON files containing an array of strings
- Expects new entries to be added sequentially at the end of the array
- No error recovery mechanisms
- Not optimized for large files
- No tests
- Many edge cases are not handled

## Usage

```bash
json-tail [flags] filename

flags:
  -i, --interval float   the interval in seconds at which to check the file for changes (default 1.0)
```

Example:
```bash
json-tail test.json -i 2.0
```

## Project Structure

```
.
├── spinner/          # Basic spinner package for terminal progress indication
│   └── spinner.go
├── main.go          # Main application code
├── go.mod           # Go module definition
└── README.md        # This file
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Disclaimer

This code is provided AS-IS and is meant for learning purposes only. Use at your own risk. No warranty or support is provided.