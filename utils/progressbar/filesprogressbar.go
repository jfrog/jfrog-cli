package progressbar

import (
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	corelog "github.com/jfrog/jfrog-cli-core/v2/utils/log"
	"github.com/jfrog/jfrog-cli-core/v2/utils/progressbar"
	ioUtils "github.com/jfrog/jfrog-client-go/utils/io"
	"github.com/jfrog/jfrog-client-go/utils/log"

	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

var terminalWidth int

type filesProgressBarManager struct {
	// A list of progress bar objects.
	bars []progressBar
	// A wait group for all progress bars.
	barsWg *sync.WaitGroup
	// A container of all external mpb bar objects to be displayed.
	container *mpb.Progress
	// A synchronization lock object.
	barsRWMutex sync.RWMutex
	// A general work indicator spinner.
	headlineBar *mpb.Bar
	// A general tasks completion indicator.
	generalProgressBar *mpb.Bar
	// A cumulative amount of tasks
	tasksCount int64
	// The log file
	logFile *os.File
}

type progressBarUnit struct {
	bar         *mpb.Bar
	incrChannel chan int
	description string
}

type progressBar interface {
	ioUtils.Progress
	getProgressBarUnit() *progressBarUnit
}

func (p *filesProgressBarManager) InitProgressReaders() {
	p.newHeadlineBar(" Working")
	p.tasksCount = 0
	p.newGeneralProgressBar()
}

// Initializes a new reader progress indicator for a new file transfer.
// Input: 'total' - file size.
//		  'label' - the title of the operation.
//		  'path' - the path of the file being processed.
// Output: progress indicator id
func (p *filesProgressBarManager) NewProgressReader(total int64, label, path string) (bar ioUtils.Progress) {
	// Write Lock when appending a new bar to the slice
	p.barsRWMutex.Lock()
	defer p.barsRWMutex.Unlock()
	p.barsWg.Add(1)
	newBar := p.container.New(int64(total),
		mpb.BarStyle().Lbound("|").Filler("🟩").Tip("🟩").Padding("⬛").Refiller("").Rbound("|"),
		mpb.BarRemoveOnComplete(),
		mpb.AppendDecorators(
			// Extra chars length is the max length of the KibiByteCounter
			decor.Name(buildProgressDescription(label, path, 17)),
			decor.CountersKibiByte("%3.1f/%3.1f"),
		),
	)

	// Add bar to bars array
	unit := initNewBarUnit(newBar, path)
	barId := len(p.bars) + 1
	readerProgressBar := ReaderProgressBar{progressBarUnit: unit, Id: barId}
	p.bars = append(p.bars, &readerProgressBar)
	return &readerProgressBar
}

// Changes progress indicator state and acts accordingly.
func (p *filesProgressBarManager) SetProgressState(id int, state string) {
	switch state {
	case "Merging":
		p.addNewMergingSpinner(id)
	}
}

// Initializes a new progress bar, that replaces the progress bar with the given replacedBarId
func (p *filesProgressBarManager) addNewMergingSpinner(replacedBarId int) {
	// Write Lock when appending a new bar to the slice
	p.barsRWMutex.Lock()
	defer p.barsRWMutex.Unlock()
	replacedBar := p.bars[replacedBarId-1].getProgressBarUnit()
	p.bars[replacedBarId-1].Abort()
	newBar := p.container.New(1,
		mpb.SpinnerStyle(createSpinnerFramesArray()...).PositionLeft(),
		mpb.AppendDecorators(
			decor.Name(buildProgressDescription("  Merging  ", replacedBar.description, 0)),
		),
	)
	// Bar replacement is a simple spinner and thus does not implement any read functionality
	unit := &progressBarUnit{bar: newBar, description: replacedBar.description}
	progressBar := SimpleProgressBar{progressBarUnit: unit, Id: replacedBarId}
	p.bars[replacedBarId-1] = &progressBar
}

func buildProgressDescription(label, path string, extraCharsLen int) string {
	separator := " | "
	// Max line length after decreasing bar width (*2 in case unicode chars with double width are used) and the extra chars
	descMaxLength := terminalWidth - (progressbar.ProgressBarWidth*2 + extraCharsLen)
	return buildDescByLimits(descMaxLength, " "+label+separator, shortenUrl(path), separator)
}

func shortenUrl(path string) string {
	if _, err := url.ParseRequestURI(path); err != nil {
		return path
	}

	semicolonIndex := strings.Index(path, ";")
	if semicolonIndex == -1 {
		return path
	}
	return path[:semicolonIndex]
}

func buildDescByLimits(descMaxLength int, prefix, path, suffix string) string {
	desc := prefix + path + suffix

	// Verify that the whole description doesn't exceed the max length
	if len(desc) <= descMaxLength {
		return desc
	}

	// If it does exceed, check if shortening the path will help (+3 is for "...")
	if len(desc)-len(path)+3 > descMaxLength {
		// Still exceeds, do not display desc
		return ""
	}

	// Shorten path from the beginning
	path = "..." + path[len(desc)-descMaxLength+3:]
	return prefix + path + suffix
}

func initNewBarUnit(bar *mpb.Bar, path string) *progressBarUnit {
	ch := make(chan int, 1000)
	unit := &progressBarUnit{bar: bar, incrChannel: ch, description: path}
	go incrBarFromChannel(unit)
	return unit
}

func incrBarFromChannel(unit *progressBarUnit) {
	// Increase bar while channel is open
	for n := range unit.incrChannel {
		unit.bar.IncrBy(n)
	}
}

func createSpinnerFramesArray() []string {
	black := "⬛"
	green := "🟩"
	spinnerFramesArray := make([]string, progressbar.ProgressBarWidth)
	for i := 1; i < progressbar.ProgressBarWidth-1; i++ {
		cur := "|" + strings.Repeat(black, i-1) + green + strings.Repeat(black, progressbar.ProgressBarWidth-2-i) + "|"
		spinnerFramesArray[i] = cur
	}
	return spinnerFramesArray
}

// Aborts a progress bar.
// Should be called even if bar completed successfully.
// The progress component's Abort method has no effect if bar has already completed, so can always be safely called anyway
func (p *filesProgressBarManager) RemoveProgress(id int) {
	p.barsRWMutex.RLock()
	defer p.barsWg.Done()
	defer p.barsRWMutex.RUnlock()
	p.bars[id-1].Abort()
	p.generalProgressBar.Increment()

}

// Quits the progress bar while aborting the initial bars.
func (p *filesProgressBarManager) Quit() (err error) {
	if p.headlineBar != nil {
		p.headlineBar.Abort(true)
		p.barsWg.Done()
		p.headlineBar = nil
	}
	if p.generalProgressBar != nil {
		p.generalProgressBar.Abort(true)
		p.barsWg.Done()
		p.generalProgressBar = nil
	}
	// Wait a refresh rate to make sure all aborts have finished
	time.Sleep(progressbar.ProgressRefreshRate)
	p.container.Wait()
	// Close the created log file (once)
	if p.logFile != nil {
		err = corelog.CloseLogFile(p.logFile)
		p.logFile = nil
		// Set back the default logger
		corelog.SetDefaultLogger()
	}
	return
}

func (p *filesProgressBarManager) GetProgress(id int) ioUtils.Progress {
	return p.bars[id-1]
}

// Initializes progress bar if possible (all conditions in 'shouldInitProgressBar' are met).
// Returns nil, nil, err if failed.
func InitFilesProgressBarIfPossible(showLogFilePath bool) (ioUtils.ProgressMgr, error) {
	shouldInit, err := progressbar.ShouldInitProgressBar()
	if !shouldInit || err != nil {
		return nil, err
	}

	logFile, err := corelog.CreateLogFile()
	if err != nil {
		return nil, err
	}
	if showLogFilePath {
		log.Info("Log path:", logFile.Name())
	}
	log.SetLogger(log.NewLogger(corelog.GetCliLogLevel(), logFile))

	newProgressBar := &filesProgressBarManager{}
	newProgressBar.barsWg = new(sync.WaitGroup)

	// Initialize the progressBar container with wg, to create a single joint point
	newProgressBar.container = mpb.New(
		mpb.WithOutput(os.Stderr),
		mpb.WithWaitGroup(newProgressBar.barsWg),
		mpb.WithWidth(progressbar.ProgressBarWidth),
		mpb.WithRefreshRate(progressbar.ProgressRefreshRate))

	newProgressBar.logFile = logFile

	return newProgressBar, nil
}

// Initializes a new progress bar for general progress indication
func (p *filesProgressBarManager) newGeneralProgressBar() {
	p.barsWg.Add(1)
	p.generalProgressBar = p.container.New(p.tasksCount,
		mpb.BarStyle().Lbound("|").Filler("⬜").Tip("⬜").Padding("⬛").Refiller("").Rbound("|"),
		mpb.BarRemoveOnComplete(),
		mpb.AppendDecorators(
			decor.Name(" Tasks: "),
			decor.CountersNoUnit("%d/%d"),
		),
	)
}

// Initializes a new progress bar for headline, with a spinner
func (p *filesProgressBarManager) newHeadlineBar(headline string) {
	p.barsWg.Add(1)
	p.headlineBar = p.container.New(1,
		mpb.SpinnerStyle("∙∙∙∙∙∙", "●∙∙∙∙∙", "∙●∙∙∙∙", "∙∙●∙∙∙", "∙∙∙●∙∙", "∙∙∙∙●∙", "∙∙∙∙∙●", "∙∙∙∙∙∙").PositionLeft(),
		mpb.BarRemoveOnComplete(),
		mpb.PrependDecorators(
			decor.Name(headline),
		),
	)
}

func (p *filesProgressBarManager) SetHeadlineMsg(msg string) {
	if p.headlineBar != nil {
		current := p.headlineBar
		p.barsRWMutex.RLock()
		// First abort, then mark progress as done and finally release the lock.
		defer p.barsRWMutex.RUnlock()
		defer p.barsWg.Done()
		defer current.Abort(true)
	}
	// Remove emojis from non-supported terminals
	msg = coreutils.RemoveEmojisIfNonSupportedTerminal(msg)
	p.newHeadlineBar(msg)
}

func (p *filesProgressBarManager) ClearHeadlineMsg() {
	if p.headlineBar != nil {
		p.barsRWMutex.RLock()
		p.headlineBar.Abort(true)
		p.barsWg.Done()
		p.barsRWMutex.RUnlock()
		// Wait a refresh rate to make sure the abort has finished
		time.Sleep(progressbar.ProgressRefreshRate)
	}
	p.headlineBar = nil
}

// IncGeneralProgressTotalBy increments the general progress bar total count by given n.
func (p *filesProgressBarManager) IncGeneralProgressTotalBy(n int64) {
	atomic.AddInt64(&p.tasksCount, n)
	if p.generalProgressBar != nil {
		p.generalProgressBar.SetTotal(p.tasksCount, false)
	}
}

type CommandWithProgress interface {
	commands.Command
	SetProgress(ioUtils.ProgressMgr)
}

func ExecWithProgress(cmd CommandWithProgress) (err error) {
	// Show log file path on all progress bars except 'setup' command
	showLogFilePath := cmd.CommandName() != "setup"
	// Init progress bar.
	progressBar, err := InitFilesProgressBarIfPossible(showLogFilePath)
	if err != nil {
		return err
	}
	if progressBar != nil {
		cmd.SetProgress(progressBar)
		defer func() {
			e := progressBar.Quit()
			if err == nil {
				err = e
			}
		}()
	}
	err = commands.Exec(cmd)
	return
}
