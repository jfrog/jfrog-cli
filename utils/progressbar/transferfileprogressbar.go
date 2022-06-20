package progressbar

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/vbauerster/mpb/v7"
	"sync/atomic"
	"time"
)

// TransferProgressMng Managing all self-hosted to SaaS transfer progress details.
// Transferring one repository's data at a time.
type TransferProgressMng struct {
	// Task bar with the total repositories transfer progress
	totalRepositories *tasksWithHeadlineProg
	// Current repo progress bars
	currentRepoHeadline *mpb.Bar
	emptyLine           *mpb.Bar
	phases              []*tasksWithHeadlineProg
	// Progress bar manager
	barsMng *ProgressBarMng
}

// NewTransferProgressMng Create TransferProgressMng object.
// If the progress bar shouldn't be displayed returns nil.
func NewTransferProgressMng(totalRepositories int64) (*TransferProgressMng, error) {
	mng, shouldDisplay, err := NewBarsMng()
	if !shouldDisplay || err != nil {
		return nil, err
	}
	transfer := TransferProgressMng{barsMng: mng}
	// Init the total repositories transfer progress bar
	transfer.totalRepositories = transfer.barsMng.NewTasksWithHeadlineProg(totalRepositories, cliutils.ColorTitle("Transferring your repositories"), true, WHITE)
	return &transfer, nil
}

// NewRepository adding new repository's progress details.
// Aborting previous repository if exists.
func (t *TransferProgressMng) NewRepository(name string, tasksPhase1, tasksPhase2 int64) {
	// Abort previous repository before creating the new one
	if t.currentRepoHeadline != nil {
		t.removeRepository()
	}
	t.currentRepoHeadline = t.barsMng.NewHeadlineBar("Current repository: "+cliutils.ColorTitle(name), false)
	t.emptyLine = t.barsMng.NewHeadlineBar("", false)
	t.addPhases(tasksPhase1, tasksPhase2)
}

// Quit terminate the TransferProgressMng process.
func (t *TransferProgressMng) Quit() {
	t.removeRepository()
	t.barsMng.quitTasksWithHeadlineProg(t.totalRepositories)
	// Wait for all go routines to finish before quiting
	t.barsMng.barsWg.Wait()
}

// IncrementPhase increments completed tasks count for a specific phase by 1.
func (t *TransferProgressMng) IncrementPhase(id int) error {
	if id < 0 || id > len(t.phases)-1 {
		return errorutils.CheckError(errors.New("invalid phase id"))
	}
	if t.phases[id].tasksProgressBar.totalTasks == 0 {
		return errorutils.CheckError(errors.New("trying to increase tasks bar that was done in previous run. "))
	}
	t.barsMng.Increment(t.phases[id])
	return nil
}

func (t *TransferProgressMng) addPhases(tasksPhase1, tasksPhase2 int64) {
	t.phases = append(t.phases, t.barsMng.NewTasksWithHeadlineProg(tasksPhase1, "Phase 1: files transfer", false, GREEN))
	t.phases = append(t.phases, t.barsMng.NewTasksWithHeadlineProg(tasksPhase2, "Phase 2: files’ diff transfer", false, GREEN))
}

func (t *TransferProgressMng) removeRepository() {
	if t.currentRepoHeadline == nil {
		return
	}
	// Increment total repositories progress bar and abort all current repo bars.
	t.barsMng.Increment(t.totalRepositories)
	t.currentRepoHeadline.Abort(true)
	t.currentRepoHeadline = nil

	t.emptyLine.Abort(true)
	t.emptyLine = nil

	// Abort all phases bars
	for i := 0; i < len(t.phases); i++ {
		t.barsMng.quitTasksWithHeadlineProg(t.phases[i])
	}
	t.phases = nil
	time.Sleep(progressRefreshRate)
}

// Progress that includes two bars:
// 1. Headline bar
// 2. Tasks counter progress bar.
type tasksWithHeadlineProg struct {
	headlineBar      *mpb.Bar
	tasksProgressBar *tasksProgressBar
	emptyLine        *mpb.Bar
}

type tasksProgressBar struct {
	bar        *mpb.Bar
	tasksCount int64
	totalTasks int64
}

// IncGeneralProgressTotalBy increments the amount of total tasks by n.
func (p *tasksProgressBar) IncGeneralProgressTotalBy(n int64) {
	atomic.AddInt64(&p.totalTasks, n)
	if p.bar != nil {
		p.bar.SetTotal(p.totalTasks, false)
	}
}

// progress bar test
func ActualTestProgressbar() (err error) {
	var total int64 = 30
	repoProg, err := NewTransferProgressMng(total)
	if err != nil {
		return err
	}
	for a := 0; a < int(total); a++ {

		repoProg.NewRepository(fmt.Sprintf("test%d", a), 0, total)
		if err != nil {
			return err
		}
		for j := 1; j < 2; j++ {
			for i := 0; i < int(total); i++ {
				time.Sleep(100000000)
				err = repoProg.IncrementPhase(j)
				if err != nil {
					return err
				}

			}
		}
		repoProg.removeRepository()

	}
	repoProg.Quit()
	return
}
