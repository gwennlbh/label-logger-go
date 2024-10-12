package labellogger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gosuri/uiprogress"
	"github.com/mitchellh/colorstring"
)

var progressbar *uiprogress.Bar
var progressBars *uiprogress.Progress

func StartProgressBar(total int, verb string, color string) {
	if progressbar != nil {
		panic("progress bar already started")
	}

	if isInteractiveTerminal() || os.Getenv("FORCE_PROGRESS_BAR") == "1" {
		Debug("terminal is interactive, starting progress bar")
	} else {
		Debug("not starting progress bar because not in an interactive terminal")
		return
	}

	progressBars = uiprogress.New()
	progressBars.SetRefreshInterval(1 * time.Millisecond)
	progressbar = progressBars.AddBar(total)
	progressbar.Empty = ' '
	progressbar.Fill = '='
	progressbar.Head = '>'
	progressbar.Width = 30
	progressbar.PrependFunc(makeProgressBarPrependFunc(verb, color))
	progressbar.AppendFunc(makeProgressBarAppendFunc(""))
	progressBars.Start()
}

func IncrementProgressBar(onDone ...func()) {
	if ProgressBarFinished() {
		if len(onDone) > 0 {
			onDone[0]()
		}
		StopProgressBar()
	}

	progressbar.Incr()
}

func ProgressBarFinished() bool {
	if progressbar == nil {
		return false
	}
	return progressbar.CompletedPercent() >= 100
}

func StopProgressBar() {
	if progressbar == nil {
		return
	}

	progressBars.Bars = nil
	progressBars.Stop()
	// Clear progress bar empty line
	fmt.Print("\r\033[K")
}

func UpdateProgressBar(verb string, color string, message string, details ...string) {
	if progressbar == nil {
		return
	}

	progressbar.PrependFunc(makeProgressBarPrependFunc(verb, color))
	progressbar.AppendFunc(makeProgressBarAppendFunc(message, details...))
}

func makeProgressBarPrependFunc(verb string, color string) func(*uiprogress.Bar) string {
	return func(b *uiprogress.Bar) string {
		if ShowingColors() {
			return colorstring.Color(
				fmt.Sprintf(`[%s][bold]%s[reset]`, color, padVerb(verb)),
			)
		}
		return padVerb(verb)
	}
}

func makeProgressBarAppendFunc(message string, details ...string) func(*uiprogress.Bar) string {
	return func(b *uiprogress.Bar) string {
		return fmt.Sprintf("%d/%d %s [dim]%s[reset]", b.Current(), b.Total, message, strings.Join(details, " "))
	}
}

func padVerb(verb string) string {
	return fmt.Sprintf("%15s", verb)
}
