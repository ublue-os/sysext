package percent

import (
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/jedib0t/go-pretty/v6/text"
	"time"
)

type Incrementer struct {
	doneIncrements int
	MaxIncrements  int
}

type IncrementTracker struct {
	Tracker     *progress.Tracker
	incrementer *Incrementer
}

var CuteColors = progress.StyleColors{
	Message: text.Colors{text.FgWhite},
	Error:   text.Colors{text.FgRed},
	Percent: text.Colors{text.FgHiBlue},
	Pinned:  text.Colors{text.BgHiBlack, text.FgWhite, text.Bold},
	Stats:   text.Colors{text.FgHiBlack},
	Time:    text.Colors{text.FgBlue},
	Tracker: text.Colors{text.FgHiBlue},
	Value:   text.Colors{text.FgBlue},
	Speed:   text.Colors{text.FgBlue},
}

func NewProgressWriter() progress.Writer {
	pw := progress.NewWriter()
	pw.SetTrackerLength(25)
	pw.Style().Visibility.TrackerOverall = true
	pw.Style().Visibility.Time = true
	pw.Style().Visibility.Tracker = true
	pw.Style().Visibility.Value = true
	pw.SetMessageLength(24)
	pw.SetSortBy(progress.SortByPercentDsc)
	pw.SetStyle(progress.StyleBlocks)
	pw.SetTrackerPosition(progress.PositionRight)
	pw.SetUpdateFrequency(time.Millisecond * 100)
	pw.Style().Options.PercentFormat = "%4.1f%%"
	pw.Style().Colors = CuteColors
	return pw
}

func NewIncrementTracker(tracker *progress.Tracker, max_increments int) *IncrementTracker {
	return &IncrementTracker{
		Tracker:     tracker,
		incrementer: &Incrementer{MaxIncrements: max_increments},
	}
}

func (it *IncrementTracker) IncrementSection() {
	var increment_step float64
	if it.incrementer.doneIncrements == 0 {
		increment_step = 10
	} else {
		increment_step = float64(it.Tracker.Total / int64(it.incrementer.MaxIncrements))
	}
	it.Tracker.Increment(int64(increment_step))
	it.incrementer.doneIncrements++
}
