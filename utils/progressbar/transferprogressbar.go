package progressbar

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

func ActualTestProgressbar() (err error) {
	var total int64 = 10
	repoProg, err := NewTransferProgressMng(100)
	if err != nil {
		return err
	}
	for a := 0; a < int(total); a++ {

		repoProg.NewRepository(fmt.Sprintf("test%d", a), total, total, total)
		if err != nil {
			return err
		}
		for j := 0; j < 3; j++ {
			for i := 0; i < int(total); i++ {
				time.Sleep(100000000)
				err = repoProg.IncrementPhase(j)
				if err != nil {
					return err
				}

			}
		}
		repoProg.RemoveRepository()

	}
	repoProg.Done()
	return
}

// TransferProgressMng represents the total progresses bars that are being shown in the terminal.
// Shows the total transfer progress details.
// For each repository that being transfer shows its specific details.
type TransferProgressMng struct {
	totalRepositories *tasksWithHeadlineProg
	emptyLine         *mpb.Bar
	// Current repo progress bars
	currentRepoHeadline *mpb.Bar
	secondEmptyLine     *mpb.Bar
	phases              []*tasksWithHeadlineProg
	barsMng             *BarsMng
}

type BarsMng struct {
	// A container of all external mpb bar objects to be displayed.
	container *mpb.Progress
	// A synchronization lock object.
	barsRWMutex sync.RWMutex
}

func NewBarsMng() *BarsMng {
	p := BarsMng{}
	p.container = mpb.New(
		mpb.WithOutput(os.Stderr),
		mpb.WithWidth(longProgressBarWidth),
		mpb.WithRefreshRate(progressRefreshRate))
	return &p
}

// NewHeadlineBar Initializes a new progress bar for headline, with an optional spinner
func (bm *BarsMng) NewHeadlineBar(msg string, spinner bool) *mpb.Bar {
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

// NewTransferProgressMng Create TransferProgressMng object.
// If the progress bar should be displayed returns nil.
func NewTransferProgressMng(totalRepositories int64) (*TransferProgressMng, error) {
	// Determine whether the progress bar should be displayed
	shouldInit, err := ShouldInitProgressBar()
	if !shouldInit || err != nil {
		return nil, err
	}
	transfer := TransferProgressMng{barsMng: NewBarsMng()}
	// Init total repositories progress bar
	transfer.totalRepositories = transfer.barsMng.NewTasksWithHeadlineProg(totalRepositories, cliutils.ColorTitle("Transferring your repositories"), true)
	transfer.emptyLine = transfer.barsMng.NewHeadlineBar("", false)
	return &transfer, nil
}

func (t *TransferProgressMng) NewRepository(name string, tasksPhase1, tasksPhase2, tasksPhase3 int64) {
	t.currentRepoHeadline = t.barsMng.NewHeadlineBar("Current repository: "+cliutils.ColorTitle(name), false)
	t.secondEmptyLine = t.barsMng.NewHeadlineBar("", false)
	t.addPhases(tasksPhase1, tasksPhase2, tasksPhase3)
}

func (t *TransferProgressMng) Done() {
	t.totalRepositories.headlineBar.Abort(false)
	t.totalRepositories.tasksProgressBar.bar.Abort(false)
	t.emptyLine.Abort(true)
}

func (t *TransferProgressMng) RemoveRepository() {
	t.currentRepoHeadline.Abort(true)
	t.currentRepoHeadline = nil
	t.secondEmptyLine.Abort(true)
	t.secondEmptyLine = nil
	// Abort all phases bars
	for i := 0; i < len(t.phases); i++ {
		t.phases[i].headlineBar.Abort(true)
		t.phases[i].tasksProgressBar.bar.Abort(true)
	}
	t.phases = nil
	t.barsMng.Increment(t.totalRepositories)
	time.Sleep(progressRefreshRate)
}

// IncrementPhase increments completed tasks count for a specific phase by 1.
func (t *TransferProgressMng) IncrementPhase(id int) error {
	if id < 0 || id > len(t.phases)-1 {
		return errorutils.CheckError(errors.New("invalid phase id"))
	}
	t.barsMng.Increment(t.phases[id])
	return nil
}

func (t *TransferProgressMng) addPhases(tasksPhase1, tasksPhase2, tasksPhase3 int64) {
	t.phases = append(t.phases, t.barsMng.NewTasksWithHeadlineProg(tasksPhase1, fmt.Sprintf("Phase 1: files transfer"), false))
	t.phases = append(t.phases, t.barsMng.NewTasksWithHeadlineProg(tasksPhase2, fmt.Sprintf("Phase 2: files transfer"), false))
	t.phases = append(t.phases, t.barsMng.NewTasksWithHeadlineProg(tasksPhase3, fmt.Sprintf("Phase 3: files transfer"), false))
}

// Progress that includes two bars:
// 1. Headline bar
// 2. Tasks counter progress bar.
type tasksWithHeadlineProg struct {
	headlineBar      *mpb.Bar
	tasksProgressBar *tasksProgressBar
}

type tasksProgressBar struct {
	bar        *mpb.Bar
	tasksCount int64
	totalTasks int64
}

func (bm *BarsMng) NewTasksWithHeadlineProg(totalTasks int64, headline string, spinner bool) *tasksWithHeadlineProg {
	prog := tasksWithHeadlineProg{}
	prog.headlineBar = bm.NewHeadlineBar(headline, spinner)
	prog.tasksProgressBar = bm.NewTasksProgressBar(totalTasks)
	return &prog
}

// Increment increments completed tasks count by 1.
func (bm *BarsMng) Increment(prog *tasksWithHeadlineProg) {
	bm.barsRWMutex.RLock()
	defer bm.barsRWMutex.RUnlock()
	prog.tasksProgressBar.bar.Increment()
	prog.tasksProgressBar.tasksCount++
	if prog.tasksProgressBar.totalTasks == prog.tasksProgressBar.tasksCount {
		// TODO:try to add emoji
	}
}

// IncGeneralProgressTotalBy increments the amount of total tasks by n.
func (p *tasksProgressBar) IncGeneralProgressTotalBy(n int64) {
	atomic.AddInt64(&p.totalTasks, n)
	if p.bar != nil {
		p.bar.SetTotal(p.totalTasks, false)
	}
}

func (bm *BarsMng) NewTasksProgressBar(totalTasks int64) *tasksProgressBar {
	pb := &tasksProgressBar{}
	pb.bar = bm.container.Add(0,
		mpb.NewBarFiller(mpb.BarStyle().Lbound("|").Filler("⬜").Tip("⬜").Padding("⬛").Refiller("").Rbound("|")),
		mpb.BarRemoveOnComplete(),
		mpb.AppendDecorators(
			decor.Name(" Tasks: "),
			decor.CountersNoUnit("%d/%d"),
		),
	)
	pb.IncGeneralProgressTotalBy(totalTasks)
	return pb
}
