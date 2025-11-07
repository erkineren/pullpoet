package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

// UI provides rich terminal UI functionality
type UI struct {
	colors       bool
	progressBars bool
	emoji        bool
	verbose      bool
	theme        string
	output       io.Writer
}

// Config holds UI configuration
type Config struct {
	Colors       bool
	ProgressBars bool
	Emoji        bool
	Verbose      bool
	Theme        string
}

// DefaultConfig returns default UI configuration
func DefaultConfig() Config {
	return Config{
		Colors:       true,
		ProgressBars: true,
		Emoji:        true,
		Verbose:      false,
		Theme:        "auto",
	}
}

// New creates a new UI instance
func New(config Config) *UI {
	// Auto-detect if colors should be disabled
	colors := config.Colors
	if config.Theme == "auto" {
		// Disable colors if not a terminal or NO_COLOR is set
		if os.Getenv("NO_COLOR") != "" || !isTerminal(os.Stdout) {
			colors = false
		}
	}

	return &UI{
		colors:       colors,
		progressBars: config.ProgressBars,
		emoji:        config.Emoji,
		verbose:      config.Verbose,
		theme:        config.Theme,
		output:       os.Stdout,
	}
}

// isTerminal checks if the writer is a terminal
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		stat, err := f.Stat()
		if err != nil {
			return false
		}
		return (stat.Mode() & os.ModeCharDevice) != 0
	}
	return false
}

// Section prints a section header
func (ui *UI) Section(title string) {
	if ui.colors {
		color.New(color.Bold, color.FgCyan).Fprintf(ui.output, "\n%s %s\n", ui.getEmoji("📋"), title)
		color.New(color.FgCyan).Fprintln(ui.output, strings.Repeat("═", len(title)+2))
	} else {
		fmt.Fprintf(ui.output, "\n%s %s\n", ui.getEmoji(""), title)
		fmt.Fprintln(ui.output, strings.Repeat("=", len(title)+2))
	}
}

// Success prints a success message
func (ui *UI) Success(message string) {
	if ui.colors {
		color.New(color.FgGreen).Fprintf(ui.output, "%s %s\n", ui.getEmoji("✅"), message)
	} else {
		fmt.Fprintf(ui.output, "%s %s\n", ui.getEmoji("[OK]"), message)
	}
}

// Info prints an info message
func (ui *UI) Info(message string) {
	if ui.colors {
		color.New(color.FgCyan).Fprintf(ui.output, "%s %s\n", ui.getEmoji("ℹ️"), message)
	} else {
		fmt.Fprintf(ui.output, "%s %s\n", ui.getEmoji("[i]"), message)
	}
}

// Warning prints a warning message
func (ui *UI) Warning(message string) {
	if ui.colors {
		color.New(color.FgYellow).Fprintf(ui.output, "%s %s\n", ui.getEmoji("⚠️"), message)
	} else {
		fmt.Fprintf(ui.output, "%s %s\n", ui.getEmoji("[!]"), message)
	}
}

// Error prints an error message
func (ui *UI) Error(message string) {
	if ui.colors {
		color.New(color.FgRed).Fprintf(ui.output, "%s %s\n", ui.getEmoji("❌"), message)
	} else {
		fmt.Fprintf(ui.output, "%s %s\n", ui.getEmoji("[ERROR]"), message)
	}
}

// Step prints a step message with indentation
func (ui *UI) Step(message string) {
	if ui.colors {
		color.New(color.FgBlue).Fprintf(ui.output, "   %s %s\n", ui.getEmoji("▸"), message)
	} else {
		fmt.Fprintf(ui.output, "   %s %s\n", ui.getEmoji(">"), message)
	}
}

// Verbose prints a message only if verbose mode is enabled
func (ui *UI) Verbose(message string) {
	if ui.verbose {
		if ui.colors {
			color.New(color.Faint).Fprintf(ui.output, "   %s %s\n", ui.getEmoji("🔍"), message)
		} else {
			fmt.Fprintf(ui.output, "   [DEBUG] %s\n", message)
		}
	}
}

// Progress creates a new progress bar
func (ui *UI) Progress(total int, description string) *ProgressBar {
	if !ui.progressBars || total <= 0 {
		return &ProgressBar{ui: ui, enabled: false}
	}

	bar := progressbar.NewOptions(total,
		progressbar.OptionEnableColorCodes(ui.colors),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetWidth(40),
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWriter(ui.output),
		progressbar.OptionClearOnFinish(),
	)

	return &ProgressBar{
		ui:      ui,
		bar:     bar,
		enabled: true,
		total:   total,
	}
}

// Spinner creates a new spinner
func (ui *UI) Spinner(message string) *Spinner {
	return &Spinner{
		ui:      ui,
		message: message,
		active:  false,
	}
}

// getEmoji returns emoji if enabled, otherwise returns fallback
func (ui *UI) getEmoji(emoji string) string {
	if ui.emoji && emoji != "" {
		return emoji
	}
	return ""
}

// Print prints a message without any formatting
func (ui *UI) Print(message string) {
	fmt.Fprintln(ui.output, message)
}

// Printf prints a formatted message
func (ui *UI) Printf(format string, args ...interface{}) {
	fmt.Fprintf(ui.output, format, args...)
}

// ProgressBar wraps progressbar functionality
type ProgressBar struct {
	ui      *UI
	bar     *progressbar.ProgressBar
	enabled bool
	total   int
	current int
}

// Add increments the progress bar
func (pb *ProgressBar) Add(n int) {
	if !pb.enabled {
		return
	}
	pb.current += n
	pb.bar.Add(n)
}

// Set sets the current progress
func (pb *ProgressBar) Set(n int) {
	if !pb.enabled {
		return
	}
	pb.current = n
	pb.bar.Set(n)
}

// Finish completes the progress bar
func (pb *ProgressBar) Finish() {
	if !pb.enabled {
		return
	}
	pb.bar.Finish()
}

// Describe updates the progress bar description
func (pb *ProgressBar) Describe(description string) {
	if !pb.enabled {
		return
	}
	pb.bar.Describe(description)
}

// Spinner provides a simple spinner animation
type Spinner struct {
	ui      *UI
	message string
	active  bool
	done    chan bool
}

// Start starts the spinner
func (s *Spinner) Start() {
	if !s.ui.progressBars {
		s.ui.Info(s.message)
		return
	}

	s.active = true
	s.done = make(chan bool)

	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		if !s.ui.emoji {
			frames = []string{"|", "/", "-", "\\"}
		}

		i := 0
		for {
			select {
			case <-s.done:
				// Clear the spinner line
				fmt.Fprintf(s.ui.output, "\r%s\r", strings.Repeat(" ", len(s.message)+10))
				return
			default:
				if s.ui.colors {
					color.New(color.FgCyan).Fprintf(s.ui.output, "\r%s %s", frames[i], s.message)
				} else {
					fmt.Fprintf(s.ui.output, "\r%s %s", frames[i], s.message)
				}
				i = (i + 1) % len(frames)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

// Stop stops the spinner
func (s *Spinner) Stop() {
	if !s.ui.progressBars || !s.active {
		return
	}
	s.active = false
	close(s.done)
	time.Sleep(50 * time.Millisecond) // Give time for goroutine to clean up
}

// UpdateMessage updates the spinner message
func (s *Spinner) UpdateMessage(message string) {
	s.message = message
}

// Table prints data in a table format
func (ui *UI) Table(headers []string, rows [][]string) {
	if len(rows) == 0 {
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	ui.printTableRow(headers, widths, true)
	ui.printTableDivider(widths)

	// Print rows
	for _, row := range rows {
		ui.printTableRow(row, widths, false)
	}
}

func (ui *UI) printTableRow(cells []string, widths []int, header bool) {
	var parts []string
	for i, cell := range cells {
		if i < len(widths) {
			parts = append(parts, fmt.Sprintf("%-*s", widths[i], cell))
		}
	}
	line := "  " + strings.Join(parts, " | ")

	if header && ui.colors {
		color.New(color.Bold, color.FgCyan).Fprintln(ui.output, line)
	} else {
		fmt.Fprintln(ui.output, line)
	}
}

func (ui *UI) printTableDivider(widths []int) {
	var parts []string
	for _, w := range widths {
		parts = append(parts, strings.Repeat("-", w))
	}
	fmt.Fprintf(ui.output, "  %s\n", strings.Join(parts, "-+-"))
}
