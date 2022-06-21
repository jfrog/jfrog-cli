package progressbar

import (
	corelog "github.com/jfrog/jfrog-cli-core/v2/utils/log"
	logUtils "github.com/jfrog/jfrog-cli/utils/log"
	"github.com/jfrog/jfrog-client-go/utils/log"
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
	// The log file
	logFile *os.File
}

func NewBarsMng() (mng *ProgressBarMng, shouldInit bool, err error) {
	// Determine whether the progress bar should be displayed or not
	shouldInit, err = ShouldInitProgressBar()
	if !shouldInit || err != nil {
		return
	}
	mng = &ProgressBarMng{}
	// Init log file
	mng.logFile, err = logUtils.CreateLogFile()
	if err != nil {
		return
	}
	log.Info("Log path:", mng.logFile.Name())
	log.SetLogger(log.NewLogger(corelog.GetCliLogLevel(), mng.logFile))

	mng.barsWg = new(sync.WaitGroup)
	mng.container = mpb.New(
		mpb.WithOutput(os.Stderr),
		mpb.WithWidth(longProgressBarWidth),
		mpb.WithWaitGroup(mng.barsWg),
		mpb.WithRefreshRate(progressRefreshRate))
	return
}

func (bm *ProgressBarMng) NewTasksWithHeadlineProg(totalTasks int64, headline string, spinner bool, color Color) *tasksWithHeadlineProg {
	bm.barsWg.Add(1)
	prog := tasksWithHeadlineProg{}
	if spinner {
		prog.headlineBar = bm.NewHeadlineBarWithSpinner(headline)
	} else {
		prog.headlineBar = bm.NewHeadlineBar(headline)
	}

	// If totalTasks is 0 - phase is already finished in previous run.
	if totalTasks == 0 {
		prog.tasksProgressBar = bm.NewDoneTasksProgressBar()
	} else {
		prog.tasksProgressBar = bm.NewTasksProgressBar(totalTasks, color)
	}
	prog.emptyLine = bm.NewHeadlineBar("")
	return &prog
}

func (bm *ProgressBarMng) quitTasksWithHeadlineProg(prog *tasksWithHeadlineProg) {
	prog.headlineBar.Abort(true)
	prog.tasksProgressBar.bar.Abort(true)
	prog.emptyLine.Abort(true)
	bm.barsWg.Done()
}

// NewHeadlineBar Initializes a new progress bar for headline, with an optional spinner
func (bm *ProgressBarMng) NewHeadlineBarWithSpinner(msg string) *mpb.Bar {
	return bm.container.New(1,
		mpb.SpinnerStyle("∙∙∙∙∙∙", "●∙∙∙∙∙", "∙●∙∙∙∙", "∙∙●∙∙∙", "∙∙∙●∙∙", "∙∙∙∙●∙", "∙∙∙∙∙●", "∙∙∙∙∙∙").PositionLeft(),
		mpb.BarRemoveOnComplete(),
		mpb.PrependDecorators(
			decor.Name(msg),
		),
	)
}

func (bm *ProgressBarMng) NewHeadlineBar(msg string) *mpb.Bar {
	return bm.container.Add(1,
		nil,
		mpb.BarRemoveOnComplete(),
		mpb.PrependDecorators(
			decor.Name(msg),
		),
	)
}

// Increment increments completed tasks count by 1.
func (bm *ProgressBarMng) Increment(prog *tasksWithHeadlineProg) {
	bm.barsRWMutex.RLock()
	defer bm.barsRWMutex.RUnlock()
	prog.tasksProgressBar.bar.Increment()
	prog.tasksProgressBar.tasksCount++
}

func (bm *ProgressBarMng) NewTasksProgressBar(totalTasks int64, color Color) *tasksProgressBar {
	pb := &tasksProgressBar{}
	filter := filterColor(color)
	pb.bar = bm.container.New(0,
		mpb.BarStyle().Lbound("|").Filler(filter).Tip(filter).Padding("⬛").Refiller("").Rbound("|"),
		mpb.BarRemoveOnComplete(),
		mpb.AppendDecorators(
			decor.Name(" Tasks: "),
			decor.CountersNoUnit("%d/%d"),
		),
	)
	pb.IncGeneralProgressTotalBy(totalTasks)
	return pb
}

func (bm *ProgressBarMng) NewDoneTasksProgressBar() *tasksProgressBar {
	pb := &tasksProgressBar{}
	pb.bar = bm.container.Add(1,
		nil,
		mpb.BarRemoveOnComplete(),
		mpb.PrependDecorators(
			decor.Name("Done ✅"),
		),
	)
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
