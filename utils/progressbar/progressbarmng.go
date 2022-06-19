package progressbar

import (
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"os"
	"sync"
	"time"
)

const (
	progressBarWidth     = 20
	longProgressBarWidth = 100
	progressRefreshRate  = 200 * time.Millisecond
)

type Color int64

const (
	WHITE Color = iota
	GREEN       = 1
)

type ProgressBarMng struct {
	// A container of all external mpb bar objects to be displayed.
	container *mpb.Progress
	// A synchronization lock object.
	barsRWMutex sync.RWMutex
	// A wait group for all progress bars.
	barsWg *sync.WaitGroup
}

func NewBarsMng() *ProgressBarMng {
	p := ProgressBarMng{}
	p.barsWg = new(sync.WaitGroup)
	p.container = mpb.New(
		mpb.WithOutput(os.Stderr),
		mpb.WithWidth(longProgressBarWidth),
		mpb.WithWaitGroup(p.barsWg),
		mpb.WithRefreshRate(progressRefreshRate))
	return &p
}

// NewHeadlineBar Initializes a new progress bar for headline, with an optional spinner
func (bm *ProgressBarMng) NewHeadlineBar(msg string, spinner bool) *mpb.Bar {
	var filter mpb.BarFiller
	if spinner {
		filter = mpb.NewBarFiller(mpb.SpinnerStyle("∙∙∙∙∙∙", "●∙∙∙∙∙", "∙●∙∙∙∙", "∙∙●∙∙∙", "∙∙∙●∙∙", "∙∙∙∙●∙", "∙∙∙∙∙●", "∙∙∙∙∙∙").PositionLeft())
	}
	headlineBar := bm.container.Add(1,
		filter,
		mpb.BarRemoveOnComplete(),
		mpb.PrependDecorators(
			decor.Name(msg),
		),
	)
	return headlineBar
}

func (bm *ProgressBarMng) NewTasksWithHeadlineProg(totalTasks int64, headline string, spinner bool, color Color) *tasksWithHeadlineProg {
	bm.barsWg.Add(1)
	prog := tasksWithHeadlineProg{}
	prog.headlineBar = bm.NewHeadlineBar(headline, spinner)
	prog.tasksProgressBar = bm.NewTasksProgressBar(totalTasks, color)
	prog.emptyLine = bm.NewHeadlineBar("", false)
	return &prog
}

func (bm *ProgressBarMng) quitTasksWithHeadlineProg(prog *tasksWithHeadlineProg) {
	prog.headlineBar.Abort(true)
	prog.tasksProgressBar.bar.Abort(true)
	prog.emptyLine.Abort(true)
	bm.barsWg.Done()
}

// Increment increments completed tasks count by 1.
func (bm *ProgressBarMng) Increment(prog *tasksWithHeadlineProg) {
	bm.barsRWMutex.RLock()
	defer bm.barsRWMutex.RUnlock()
	prog.tasksProgressBar.bar.Increment()
	prog.tasksProgressBar.tasksCount++
	if prog.tasksProgressBar.totalTasks == prog.tasksProgressBar.tasksCount {
		// TODO: done wait
	}
}

func (bm *ProgressBarMng) NewTasksProgressBar(totalTasks int64, color Color) *tasksProgressBar {
	pb := &tasksProgressBar{}
	filter := filterColor(color)
	pb.bar = bm.container.Add(0,
		mpb.NewBarFiller(mpb.BarStyle().Lbound("|").Filler(filter).Tip(filter).Padding("⬛").Refiller("").Rbound("|")),
		mpb.BarRemoveOnComplete(),
		mpb.AppendDecorators(
			decor.Name(" Tasks: "),
			decor.CountersNoUnit("%d/%d"),
		),
	)
	pb.IncGeneralProgressTotalBy(totalTasks)
	return pb
}

func filterColor(color Color) (filter string) {
	switch color {
	case GREEN:
		filter = "🟩"
	case WHITE:
		filter = "⬜"
	default:
		filter = "⬜"
	}
	return
}
