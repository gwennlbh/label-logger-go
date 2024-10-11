package labellogger

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/mitchellh/colorstring"
)

func logWriter() io.Writer {
	var writer io.Writer = os.Stderr
	if progressBars != nil {
		writer = progressBars.Bypass()
	}
	return writer
}

func indentSubsequent(size int, text string) string {
	indentation := strings.Repeat(" ", size)
	return strings.ReplaceAll(text, "\n", "\n"+indentation)
}

// Log logs a custom message with a verb and color.
func Log(verb string, color string, message string, fmtArgs ...interface{}) {
	fmt.Fprintln(logWriter(), colorstring.Color(fmt.Sprintf("[bold][%s]%15s[reset] %s", color, verb, indentSubsequent(15+1, fmt.Sprintf(message, fmtArgs...)))))
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

// Warn logs warnings.
func Warn(message string, fmtArgs ...interface{}) {
	Log("Warning", "yellow", message, fmtArgs...)
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
