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

		repoProg.NewRepositoryProgressMng(fmt.Sprintf("test%d", a), total, total, total)
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
		repoProg.RemoveRepositoryProgressMng()
		repoProg.totalRepositories.Increment()
	}
	return
}

// TransferProgress represents the total progresses bars that are being shown in the terminal.
// Shows the total transfer progress details.
// For each repository that being transfer shows its specific details.
type TransferProgress struct {
	totalRepositories *tasksWithHeadlineProg
	// Progress of the repository that currently being transferred
	currentRepo *currentRepoProgressMng
	// A container of all external mpb bar objects to be displayed.
	container *mpb.Progress
}

func NewTransferProgress(totalRepositories int64) *TransferProgress {
	transfer := TransferProgress{}
	// Set container
	transfer.container = mpb.New(
		mpb.WithOutput(os.Stderr),
		//mpb.WithWaitGroup(p.barsWg),
		mpb.WithWidth(progressBarWidth),
		mpb.WithRefreshRate(progressRefreshRate))

	transfer.totalRepositories = newTasksWithHeadlineProg(totalRepositories, "Transferring your Artifactory", false, transfer.container)
	return &transfer
}

func (t *TransferProgress) NewRepositoryProgressMng(name string, tasksPhase1, tasksPhase2, tasksPhase3 int64) error {
	// Determine whether the progress bar should be displayed
	shouldInit, err := ShouldInitProgressBar()
	if !shouldInit || err != nil {
		return err
	}
	t.currentRepo = &currentRepoProgressMng{}
	t.currentRepo.container = t.container
	t.currentRepo.repoName = name
	t.currentRepo.headlineBar = newHeadlineBar("Transfer repository:"+name, true, t.container)
	t.currentRepo.addPhases(tasksPhase1, tasksPhase2, tasksPhase3)
	// TODO: quit
	return nil
}

// TODO: add more
func (t *TransferProgress) RemoveRepositoryProgressMng() error {
	t.currentRepo.headlineBar.bar.Abort(true)
	// Abort all phases bars
	for i := 0; i < len(t.currentRepo.phases); i++ {
		t.currentRepo.phases[i].headlineBar.bar.Abort(true)
		t.currentRepo.phases[i].tasksProgressBar.bar.Abort(true)
		//p.barsWg.Done()
		t.currentRepo.headlineBar = nil
	}

	return nil
}

// IncrementPhase increments completed tasks count for a specific phase by 1.
func (t *TransferProgress) IncrementPhase(id int) error {
	if id < 0 || id > len(t.currentRepo.phases)-1 {
		return errorutils.CheckError(errors.New("invalid phase id"))
	}
	t.currentRepo.phases[id].Increment()
	return nil
}

// currentRepoProgressMng progressbar manager for a single repository transfer.
type currentRepoProgressMng struct {
	repoName    string
	headlineBar *headlineBar
	// Repository's transfer phases. Transfer includes 3 phases.
	phases []*phase
	// A synchronization lock object.
	barsRWMutex sync.RWMutex
	container   *mpb.Progress
}

func (barMng *currentRepoProgressMng) addPhases(tasksPhase1, tasksPhase2, tasksPhase3 int64) {
	barMng.phases = append(barMng.phases, newPhase(tasksPhase1, fmt.Sprintf("Phase 1: files transfer"), barMng.container))
	barMng.phases = append(barMng.phases, newPhase(tasksPhase2, fmt.Sprintf("Phase 2: files’ diff transfer"), barMng.container))
	barMng.phases = append(barMng.phases, newPhase(tasksPhase3, fmt.Sprintf("Phase 3: properties diff transfer"), barMng.container))
}

// phase represents a specific repository transfer phase
type phase struct {
	*tasksWithHeadlineProg
}

func newPhase(totalTasks int64, headline string, container *mpb.Progress) *phase {
	return &phase{
		tasksWithHeadlineProg: newTasksWithHeadlineProg(totalTasks, headline, false, container),
	}
}

// Progress that includes two bars:
// 1. Headline bar
// 2. Tasks counter progress bar.
type tasksWithHeadlineProg struct {
	headlineBar      *headlineBar
	tasksProgressBar *tasksProgressBar
	// A container of all external mpb bar objects to be displayed.
	container *mpb.Progress
}

type tasksProgressBar struct {
	bar *mpb.Bar
	// A container of all external mpb bar objects to be displayed.
	container  *mpb.Progress
	tasksCount int64
	totalTasks int64
}

// We can't print to terminal when using progressbar.
// headlineBar is a bar that its peruse is to show text in terminal.
type headlineBar struct {
	//string? headline
	bar *mpb.Bar
	// A container of all external mpb bar objects to be displayed.
	container *mpb.Progress
}

func newTasksWithHeadlineProg(totalTasks int64, headline string, spinner bool, container *mpb.Progress) *tasksWithHeadlineProg {
	p := tasksWithHeadlineProg{}
	//p.barsWg = new(sync.WaitGroup)
	// Initialize the progressBar container with wg, to create a single joint point
	p.container = container
	p.setNewHeadlineMsg(headline, spinner)
	p.setNewTasksProgressBar(totalTasks)
	return &p
}

func (p *tasksWithHeadlineProg) setNewHeadlineMsg(msg string, spinner bool) {
	p.headlineBar = newHeadlineBar(msg, spinner, p.container)
}

func (p *tasksWithHeadlineProg) setNewTasksProgressBar(totalTasks int64) {
	p.tasksProgressBar = NewTasksProgressBar(totalTasks, p.container)
	p.IncGeneralProgressTotalBy(p.tasksProgressBar.totalTasks)
}

// Increment increments completed tasks count by 1.
func (p *tasksWithHeadlineProg) Increment() {
	p.tasksProgressBar.bar.Increment()
	if p.tasksProgressBar.totalTasks == p.tasksProgressBar.tasksCount {
		// TODO: instead of complete try to switch to green v emoji
		p.tasksProgressBar.bar.EnableTriggerComplete()
	}
}

// IncGeneralProgressTotalBy increments the amount of total tasks by n.
func (p *tasksWithHeadlineProg) IncGeneralProgressTotalBy(n int64) {
	atomic.AddInt64(&p.tasksProgressBar.tasksCount, n)
	if p.tasksProgressBar != nil {
		p.tasksProgressBar.bar.SetTotal(p.tasksProgressBar.tasksCount, false)
	}
}

func NewTasksProgressBar(totalTasks int64, container *mpb.Progress) *tasksProgressBar {
	//p.barsWg.Add(1)
	pb := &tasksProgressBar{totalTasks: totalTasks, tasksCount: 0, container: container}
	pb.bar = pb.container.Add(pb.tasksCount,
		mpb.NewBarFiller(mpb.BarStyle().Lbound("|").Filler("⬜").Tip("⬜").Padding("⬛").Refiller("").Rbound("|")),
		//mpb.BarFillerOnComplete("✅"),
		mpb.BarRemoveOnComplete(),
		mpb.AppendDecorators(
			decor.Name(" Tasks: "),
			decor.CountersNoUnit("%d/%d"),
		),
	)
	return pb
}

// TODO duplicate
// Initializes a new progress bar for headline, with a spinner
func newHeadlineBar(headline string, spinner bool, container *mpb.Progress) *headlineBar {
	//p.barsWg.Add(1)
	headlineBar := headlineBar{}
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
