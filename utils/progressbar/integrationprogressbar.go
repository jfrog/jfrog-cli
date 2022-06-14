package progressbar

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"

	//ioUtils "github.com/jfrog/jfrog-client-go/utils/io"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

func ActualTestProgressbar() (err error) {
	var total int64 = 5
	repoProg := NewTransferProgress(10)
	for a := 0; a < 6; a++ {

		repoProg.AddRepository(fmt.Sprintf("test%d", a), total, total, total)
		if err != nil {
			return err
		}
		for j := 0; j < 3; j++ {
			for i := 0; i < int(total); i++ {
				time.Sleep(1000000000)
				err = repoProg.IncrementPhase(j)
				if err != nil {
					return err
				}

			}
		}
		repoProg.RemoveRepository()
		repoProg.generalBar.Increment()
	}
	return
}

// TransferProgress represents the total progresses bars that are being shown in the terminal.
// Shows the total transfer progress details.
// For each repository that being transfer shows its specific details.
type TransferProgress struct {
	generalBar *tasksProgressbarWithHeadline

	// A container of all external mpb bar objects to be displayed.
	container *mpb.Progress
	// Progress of the repository that currently being transferred
	repoProgressMng
}

func NewTransferProgress(totalRepositories int64) *TransferProgress {
	transfer := TransferProgress{}
	transfer.container = mpb.New(
		mpb.WithOutput(os.Stderr),
		//mpb.WithWaitGroup(p.barsWg),
		mpb.WithWidth(progressBarWidth),
		mpb.WithRefreshRate(progressRefreshRate))
	transfer.generalBar = newTasksProgressbarWithHeadline(totalRepositories, "Transferring your Artifactory", false, transfer.container)
	return &transfer
}

func (t *TransferProgress) AddRepository(repoName string, tasksPhase1, tasksPhase2, tasksPhase3 int64) error {
	currentRepo, err := t.newRepositoryProgress(repoName, tasksPhase1, tasksPhase2, tasksPhase3)
	if err != nil {
		return err
	}
	t.phases = currentRepo.phases
	t.headlineBar = currentRepo.headlineBar
	return nil
}

// RemoveRepository TODO
func (t *TransferProgress) RemoveRepository() error {

	t.headlineBar.bar.Abort(true)
	for i := 0; i < len(t.phases); i++ {
		t.phases[i].headlineBar.bar.Abort(true)
		t.phases[i].tasksProgressBar.bar.Abort(true)
		//p.barsWg.Done()
		t.headlineBar = nil
	}

	return nil
}

// repoProgressMng progressbar manager for a single repository transfer.
type repoProgressMng struct {
	repoName    string
	headlineBar *headlineProgressBar
	// Repository's transfer phases. Transfer includes 3 phases.
	phases []*phase
	// A synchronization lock object.
	barsRWMutex sync.RWMutex
	container   *mpb.Progress
}

type phase struct {
	*tasksProgressbarWithHeadline
}

type generalTasksProgressBar struct {
	bar *mpb.Bar
	// A container of all external mpb bar objects to be displayed.
	container  *mpb.Progress
	tasksCount int64
	totalTasks int64
}

// We can't print to terminal when using progressbar.
// headlineBar is a bar that its peruse is to show text in terminal.
type headlineProgressBar struct {
	//string? headline
	bar *mpb.Bar
	// A container of all external mpb bar objects to be displayed.
	container *mpb.Progress
}

type tasksProgressbarWithHeadline struct {
	headlineBar *headlineProgressBar
	// A general tasks completion indicator.
	tasksProgressBar *generalTasksProgressBar
	// A container of all external mpb bar objects to be displayed.
	container *mpb.Progress
	// TODO
	// A wait group for all progress bars.
	//barsWg *sync.WaitGroup
}

func (p *tasksProgressbarWithHeadline) SetHeadlineMsg(msg string, spinner bool) {
	p.headlineBar = newHeadlineBar(msg, spinner, p.container)
}

func newPhase(totalTasks int64, headline string, container *mpb.Progress) *phase {
	return &phase{
		tasksProgressbarWithHeadline: newTasksProgressbarWithHeadline(totalTasks, headline, false, container),
	}
}

func newTasksProgressbarWithHeadline(totalTasks int64, headline string, spinner bool, container *mpb.Progress) *tasksProgressbarWithHeadline {
	p := tasksProgressbarWithHeadline{}
	//p.barsWg = new(sync.WaitGroup)
	// Initialize the progressBar container with wg, to create a single joint point
	p.container = container
	p.SetHeadlineMsg(headline, spinner)
	p.setTaskProgressBar(totalTasks)
	return &p
}

func (p *tasksProgressbarWithHeadline) setTaskProgressBar(totalTasks int64) {
	p.tasksProgressBar = NewGeneralTasksProgressBar(totalTasks, p.container)
	p.IncGeneralProgressTotalBy(p.tasksProgressBar.totalTasks)
}

// Initializes a new progress bar for general progress indication
func NewGeneralTasksProgressBar(totalTasks int64, container *mpb.Progress) *generalTasksProgressBar {
	//p.barsWg.Add(1)
	pb := &generalTasksProgressBar{totalTasks: totalTasks, tasksCount: 0, container: container}
	pb.bar = pb.container.Add(pb.tasksCount,
		mpb.NewBarFiller(mpb.BarStyle().Lbound("|").Filler("⬜").Tip("⬜").Padding("⬛").Refiller("").Rbound("|")),
		mpb.BarFillerOnComplete("✅"),
		mpb.AppendDecorators(
			decor.Name(" Tasks: "),
			decor.CountersNoUnit("%d/%d"),
		),
	)
	return pb
}

func (t *TransferProgress) IncrementPhase(id int) error {
	if id < 0 || id > len(t.phases)-1 {
		return errorutils.CheckError(errors.New("invalid phase id"))
	}
	t.phases[id].Increment()
	return nil
}
func (p *tasksProgressbarWithHeadline) Increment() {
	p.tasksProgressBar.bar.Increment()
	if p.tasksProgressBar.totalTasks == p.tasksProgressBar.tasksCount {
		p.tasksProgressBar.bar.EnableTriggerComplete()
	}
}

// IncGeneralProgressTotalBy incremenates the general progress bar total count by given n.
func (p *tasksProgressbarWithHeadline) IncGeneralProgressTotalBy(n int64) {
	atomic.AddInt64(&p.tasksProgressBar.tasksCount, n)
	if p.tasksProgressBar != nil {
		p.tasksProgressBar.bar.SetTotal(p.tasksProgressBar.tasksCount, false)
	}
}

// Initializes a new progress bar for headline, with a spinner
func newHeadlineBar(headline string, spinner bool, container *mpb.Progress) *headlineProgressBar {
	//p.barsWg.Add(1)
	headlineBar := headlineProgressBar{}
	var filter mpb.BarFiller
	if spinner {
		filter = mpb.NewBarFiller(mpb.SpinnerStyle("∙∙∙∙∙∙", "●∙∙∙∙∙", "∙●∙∙∙∙", "∙∙●∙∙∙", "∙∙∙●∙∙", "∙∙∙∙●∙", "∙∙∙∙∙●", "∙∙∙∙∙∙").PositionLeft())
	}
	headlineBar.bar = container.Add(1,
		filter,
		mpb.BarRemoveOnComplete(),
		mpb.PrependDecorators(
			decor.Name(headline),
		),
	)
	return &headlineBar
}

//// Initializes a new progress bar for headline, with a spinner
//func (t *TransferProgress) newHeadlineBar(headline string) {
//	//p.barsWg.Add(1)
//	t.generalBar.barheadlineBar.bar = t.container.Add(1,
//		mpb.NewBarFiller(mpb.SpinnerStyle("∙∙∙∙∙∙", "●∙∙∙∙∙", "∙●∙∙∙∙", "∙∙●∙∙∙", "∙∙∙●∙∙", "∙∙∙∙●∙", "∙∙∙∙∙●", "∙∙∙∙∙∙").PositionLeft()),
//		mpb.BarRemoveOnComplete(),
//		mpb.PrependDecorators(
//			decor.Name(headline),
//		),
//	)
//}

// Quits the progress bar while aborting the initial bars.
func (p *phase) Quit() (err error) {
	if p.headlineBar != nil {
		p.headlineBar.bar.Abort(true)
		//	p.barsWg.Done()
		p.headlineBar = nil
	}

	if p.tasksProgressBar != nil {
		p.tasksProgressBar.bar.Abort(true)
		//	p.barsWg.Done()
		p.tasksProgressBar = nil
	}
	// Wait a refresh rate to make sure all aborts have finished
	time.Sleep(progressRefreshRate)
	p.container.Wait()
	// Close the created log file (once)
	return
}

func (t *TransferProgress) newRepositoryProgress(name string, tasksPhase1, tasksPhase2, tasksPhase3 int64) (repoProgressMgr *repoProgressMng, err error) {
	// Init progress bar.
	shouldInit, err := ShouldInitProgressBar()
	if !shouldInit || err != nil {
		return nil, err
	}
	repoProgressMgr = &repoProgressMng{}
	repoProgressMgr.container = t.container
	repoProgressMgr.repoName = name
	repoProgressMgr.headlineBar = newHeadlineBar(fmt.Sprintf("\nTranfer repository: %s", name), true, t.container)
	repoProgressMgr.addPhases(tasksPhase1, tasksPhase2, tasksPhase3)
	// TODO: quit
	return
}

func (barMng *repoProgressMng) addPhases(tasksPhase1, tasksPhase2, tasksPhase3 int64) {
	barMng.phases = append(barMng.phases, newPhase(tasksPhase1, fmt.Sprintf("Phase 1: files transfer"), barMng.container))
	barMng.phases = append(barMng.phases, newPhase(tasksPhase2, fmt.Sprintf("Phase 2: files’ diff transfer"), barMng.container))
	barMng.phases = append(barMng.phases, newPhase(tasksPhase3, fmt.Sprintf("Phase 3: properties diff transfer"), barMng.container))
}
