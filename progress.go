package labellogger

import (
	"fmt"
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

	if isInteractiveTerminal() {
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
	progressbar.PrependFunc(func(b *uiprogress.Bar) string {
		return colorstring.Color(
			fmt.Sprintf(
				`[magenta][bold]%15s[reset]`,
				"Building",
			),
		)
	})
	progressbar.AppendFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("%d/%d", b.Current(), b.Total)
	})
	progressBars.Start()
}

func IncrementProgress(onDone func()) {
	if progressbar == nil {
		return
	}

	progressbar.Incr()
	if progressbar.CompletedPercent() >= 100 {
		StopProgressBar()
		onDone()
	}
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

func padVerb(verb string) string {
	return fmt.Sprintf("%15s", verb)
}

// Status updates the current progress
func UpdateProgressBar(verb string, color string, message string, details ...string) {
	formattedDetails := ""
	if len(details) > 0 {
		formattedDetails = fmt.Sprintf(" [dim]%s[reset]", strings.Join(details, " "))
	}
	formattedMessage := colorstring.Color(fmt.Sprintf("[bold][%s]%s[reset]"+formattedDetails, color, padVerb(verb), message))

	if progressBars != nil {
		fmt.Fprintln(progressBars.Bypass(), formattedMessage)
	} else {
		if isInteractiveTerminal() {
			fmt.Println(formattedMessage)
		} else {
			fmt.Printf(" %s %s %s\n", padVerb(verb), message, strings.Join(details, " "))
		}
	}
}
