package labellogger

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/mitchellh/colorstring"
)

// Set to a path to log to a file as well as stdout.
var LogFilePath string

// Set to true to prepend the date to logs.
var PrependDateToLogs = false

var showingTimingLogs = os.Getenv("DEBUG_TIMING") != ""

func logWriter(original io.Writer) io.Writer {
	writer := original
	if LogFilePath != "" {
		logfile, err := os.OpenFile(LogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			writer = io.MultiWriter(writer, logfile)
		}
	}

	if PrependDateToLogs {
		writer = prependDateWriter{out: writer}
	}

	if progressBars != nil {
		writer = progressBars.Bypass()
	}
	if !ShowingColors() {
		writer = noAnsiCodesWriter{out: writer}
	}
	return writer
}

type prependDateWriter struct {
	out io.Writer
}

func (w prependDateWriter) Write(p []byte) (n int, err error) {
	return w.out.Write([]byte(
		fmt.Sprintf("[%s] %s",
			time.Now(),
			strings.TrimLeft(string(p), " "),
		)))
}

// noAnsiCodesWriter is an io.Writer that writes to the underlying writer, but strips ANSI color codes beforehand
type noAnsiCodesWriter struct {
	out io.Writer
}

func (w noAnsiCodesWriter) Write(p []byte) (n int, err error) {
	return w.out.Write(stripansicolors(p))
}

func indentSubsequent(size int, text string) string {
	indentation := strings.Repeat(" ", size)
	return strings.ReplaceAll(text, "\n", "\n"+indentation)
}

// Log logs a custom message with a verb and color.
func Log(verb string, color string, message string, fmtArgs ...interface{}) {
	LogNoFormatting(verb, color, colorstring.Color(fmt.Sprintf(message, fmtArgs...)))
}

// LogNoColor logs a message without applying colorstring syntax to message.
func LogNoColor(verb string, color string, message string, fmtArgs ...interface{}) {
	LogNoFormatting(verb, color, fmt.Sprintf(message, fmtArgs...))
}

func LogNoFormatting(verb string, color string, message string) {
	fmt.Fprintln(
		logWriter(os.Stderr),
		colorstring.Color(fmt.Sprintf("[bold][%s]%15s[reset]", color, verb))+
			" "+
			indentSubsequent(15+1, message),
	)
}

// Error logs non-fatal errors.
func Error(message string, fmtArgs ...interface{}) {
	Log("Error", "red", message, fmtArgs...)
}

func ErrorDisplay(msg string, err error, fmtArgs ...interface{}) {
	Error(FormatErrors(fmt.Errorf(msg+": %w", append(fmtArgs, err)...)))
}

func WarnDisplay(msg string, err error, fmtArgs ...interface{}) {
	Warn(FormatErrors(fmt.Errorf(msg+": %w", append(fmtArgs, err)...)))
}

// Info logs infos.
func Info(message string, fmtArgs ...interface{}) {
	Log("Info", "blue", message, fmtArgs...)
}

// Debug logs debug information. Set DEBUG environment variable to enable.
func Debug(message string, fmtArgs ...interface{}) {
	if os.Getenv("DEBUG") == "" {
		return
	}
	Log("Debug", "magenta", message, fmtArgs...)
}

// DebugNoColor logs debug information without applying colorstring syntax to message.
func DebugNoColor(message string, fmtArgs ...interface{}) {
	if os.Getenv("DEBUG") == "" {
		return
	}
	LogNoColor("Debug", "magenta", message, fmtArgs...)
}

// Warn logs warnings.
func Warn(message string, fmtArgs ...interface{}) {
	Log("Warning", "yellow", message, fmtArgs...)
}

// Timing logs timing debug logs. Mostly used with TimeTrack. To enable, set DEBUG_TIMING environment variable.
func Timing(job string, args []interface{}, timeTaken time.Duration) {
	if !showingTimingLogs {
		return
	}
	formattedArgs := ""
	for i, arg := range args {
		if i > 0 {
			formattedArgs += " "
		}
		formattedArgs += fmt.Sprintf("%v", arg)
	}
	Log("Timing", "dim", "[bold]%-30s[reset][dim]([reset]%-50s[dim])[reset] took [yellow]%s", job, formattedArgs, timeTaken)
}

// TimeTrack logs the time taken for a function to execute, and logs out the time taken.
// Usage: at the top of your function, defer TimeTrack(time.Now(), "your job name").
// To enable, set DEBUG_TIMING environment variable.
func TimeTrack(start time.Time, job string, args ...interface{}) {
	if !showingTimingLogs {
		return
	}
	elapsed := time.Since(start)
	Timing(job, args, elapsed)
}

// List formats a list of strings with a format string and separator.
func List(list []string, format string, separator string) string {
	result := ""
	for i, tag := range list {
		sep := separator
		if i == len(list)-1 {
			sep = ""
		}
		result += fmt.Sprintf(format, tag) + sep
	}
	return result
}

// FormatErrors returns a string where the error message was split on ': ', and each item is on a new line, indented once more than the previous line.
func FormatErrors(err error) string {
	causes := strings.Split(err.Error(), ": ")
	output := ""
	for i, cause := range causes {
		output += strings.Repeat(" ", i) + cause
		if i < len(causes)-1 {
			output += "\n"
		}
	}
	return output
}

func isInteractiveTerminal() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stderr.Fd())
}

func stripansicolors(b []byte) []byte {
	// TODO find a way to do this without converting to string
	s := string(b)
	s = regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(s, "")
	return []byte(s)
}

// ShowingColors returns true if colors (ANSI escape codes) should be printed.
// Environment variables can control this: NO_COLOR=1 disables colors, and FORCE_COLOR=1 forces colors.
// Otherwise, heuristics (such as whether the output is an interactive terminal) are used.
func ShowingColors() bool {
	if os.Getenv("NO_COLOR") == "1" {
		return false
	}
	if os.Getenv("FORCE_COLOR") == "1" {
		return true
	}
	return isInteractiveTerminal()
}
