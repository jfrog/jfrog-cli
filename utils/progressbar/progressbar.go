package progressbar

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	corelog "github.com/jfrog/jfrog-cli-core/utils/log"
	logUtils "github.com/jfrog/jfrog-cli/utils/log"
	"github.com/jfrog/jfrog-client-go/utils"
	ioUtils "github.com/jfrog/jfrog-client-go/utils/io"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
	"golang.org/x/crypto/ssh/terminal"
)

var terminalWidth int

const progressBarWidth = 20
const minTerminalWidth = 70
const progressRefreshRate = 200 * time.Millisecond

type progressBarManager struct {
	bars           []progressBar
	barsWg         *sync.WaitGroup
	container      *mpb.Progress
	barsRWMutex    sync.RWMutex
	headlineBar    *mpb.Bar
	logFilePathBar *mpb.Bar
}

type progressBarUnit struct {
	bar         *mpb.Bar
	incrChannel chan int
	replaceBar  *mpb.Bar
}

type progressBar interface {
	ioUtils.ProgressBar
	getProgressBarUnit() *progressBarUnit
}

func NewProcessProgressBar(prefix, extraInformation string) (bar ioUtils.ProgressBar) {
	return nil
}

func (p *progressBarManager) NewReaderProgressBar(total int64, prefix, extraInformation string) (bar ioUtils.ProgressBar) {
	// Write Lock when appending a new bar to the slice
	p.barsRWMutex.Lock()
	p.barsWg.Add(1)
	newBar := p.container.AddBar(int64(total),
		mpb.BarStyle("⬜⬜⬜⬛⬛"),
		mpb.BarRemoveOnComplete(),
		mpb.AppendDecorators(
			// Extra chars length is the max length of the KibiByteCounter
			decor.Name(buildProgressDescription(prefix, extraInformation, 17)),
			decor.CountersKibiByte("%3.1f/%3.1f"),
		),
	)

	// Add bar to bars array
	unit := initNewBarUnit(newBar)
	barId := len(p.bars) + 1
	readerProgressBar := ReaderProgressBar{progressBarUnit: unit, Id: barId}
	p.bars = append(p.bars, &readerProgressBar)
	p.barsRWMutex.Unlock()
	return &readerProgressBar
}

// // Initializes a new progress bar
// func (p *progressBarManager) NewConsumerProgressBar(consumer parallel.Runner, prefix string) (barId int) {
// 	// Write Lock when appending a new bar to the slice
// 	p.barsRWMutex.Lock()
// 	p.barsWg.Add(1)

// 	newBar := p.container.AddBar(int64(consumer.TasksCount()),
// 		mpb.BarStyle("🐸🐸..."),
// 		mpb.BarRemoveOnComplete(),
// 		mpb.AppendDecorators(
// 			// Extra chars length is the max length of the KibiByteCounter
// 			decor.Name(buildProgressDescription(prefix, 17)),
// 			decor.CountersKibiByte("%3.1f/%3.1f"),
// 		),
// 	)

// 	// Add bar to bars array
// 	unit := initNewBarUnit(newBar)
// 	p.bars = append(p.bars, unit)
// 	barId = len(p.bars)
// 	p.barsRWMutex.Unlock()
// 	return barId
// }

// Initializes a new progress bar, that replaces an existing bar when it is completed
// Refactor it name to new replacemant spinner!
func (p *progressBarManager) AddNewReplacementSpinner(replaceBarId int, prefix, extraInformation string) (id int) {
	// Write Lock when appending a new bar to the slice
	p.barsRWMutex.Lock()
	p.barsWg.Add(1)

	newBar := p.container.AddSpinner(1, mpb.SpinnerOnMiddle,
		mpb.SpinnerStyle(createSpinnerFramesArray()),
		mpb.BarParkTo(p.bars[replaceBarId-1].getProgressBarUnit().bar),
		mpb.AppendDecorators(
			decor.Name(buildProgressDescription(prefix, extraInformation, 0)),
		),
	)

	// Bar replacement is a spinner and thus does not use a channel for incrementing
	unit := &progressBarUnit{bar: newBar, incrChannel: nil, replaceBar: p.bars[replaceBarId-1].getProgressBarUnit().bar}
	// Add bar to bars array
	barId := len(p.bars) + 1
	readerProgressBar := ReaderProgressBar{progressBarUnit: unit, Id: barId}
	p.bars = append(p.bars, &readerProgressBar)
	p.barsRWMutex.Unlock()
	return barId
}

func buildProgressDescription(prefix, path string, extraCharsLen int) string {
	separator := " | "
	// Max line length after decreasing bar width (*2 in case unicode chars with double width are used) and the extra chars
	descMaxLength := terminalWidth - (progressBarWidth*2 + extraCharsLen)
	return buildDescByLimits(descMaxLength, separator+prefix+separator, path, separator)
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

func initNewBarUnit(bar *mpb.Bar) *progressBarUnit {
	ch := make(chan int, 1000)
	unit := &progressBarUnit{bar: bar, incrChannel: ch}
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
	white := "⬜"
	spinnerFramesArray := make([]string, progressBarWidth)
	for i := 0; i < progressBarWidth; i++ {
		cur := strings.Repeat(black, i) + white + strings.Repeat(black, progressBarWidth-1-i)
		spinnerFramesArray[i] = cur
	}
	return spinnerFramesArray
}

// split this class - to prograssbarreader and mgr. make the mgr class to be responsable to the new prograssbar of the general prograss
// make sure it will be easy to add new types of prograss action like reader to move/copy.

// Aborts a progress bar.
// Should be called even if bar completed successfully.
// The progress component's Abort method has no effect if bar has already completed, so can always be safely called anyway
func (p *progressBarManager) Abort(barId int) {
	p.barsRWMutex.RLock()
	defer p.barsWg.Done()
	p.bars[barId-1].Abort()
	p.barsRWMutex.RUnlock()
}

// Quits the progress bar while aborting the initial bars.
func (p *progressBarManager) Quit() {
	if p.headlineBar != nil {
		p.headlineBar.Abort(true)
		p.barsWg.Done()
	}
	if p.logFilePathBar != nil {
		p.barsWg.Done()
	}
	// Wait a refresh rate to make sure all aborts have finished
	time.Sleep(progressRefreshRate)
	p.container.Wait()
}

func (p *progressBarManager) GetProgressBar(id int) ioUtils.ProgressBar {
	return p.bars[id-1]
}

// Initializes progress bar if possible (all conditions in 'shouldInitProgressBar' are met).
// Creates a log file and sets the Logger to it. Caller responsible to close the file.
// Returns nil, nil, err if failed.
func InitProgressBarIfPossible() (ioUtils.ProgressMgr, *os.File, error) {
	shouldInit, err := shouldInitProgressBar()
	if !shouldInit || err != nil {
		return nil, nil, err
	}

	logFile, err := logUtils.CreateLogFile()
	if err != nil {
		return nil, nil, err
	}
	log.SetLogger(log.NewLogger(corelog.GetCliLogLevel(), logFile))

	newProgressBar := &progressBarManager{}
	newProgressBar.barsWg = new(sync.WaitGroup)

	// Initialize the progressBar container with wg, to create a single joint point
	newProgressBar.container = mpb.New(
		mpb.WithOutput(os.Stderr),
		mpb.WithWaitGroup(newProgressBar.barsWg),
		mpb.WithWidth(progressBarWidth),
		mpb.WithRefreshRate(progressRefreshRate))

	// Add headline bar to the whole progress
	newProgressBar.printLogFilePathAsBar(logFile.Name())
	newProgressBar.newHeadlineBar(" Working... ")

	return newProgressBar, logFile, nil
}

// Init progress bar if all required conditions are met:
// CI == false (or unset), Stderr is a terminal, and terminal width is large enough
func shouldInitProgressBar() (bool, error) {
	ci, err := utils.GetBoolEnvValue(coreutils.CI, false)
	if ci || err != nil {
		return false, err
	}
	if !isTerminal() {
		return false, err
	}
	err = setTerminalWidthVar()
	if err != nil {
		return false, err
	}
	return terminalWidth >= minTerminalWidth, nil
}

// Check if Stderr is a terminal
func isTerminal() bool {
	return terminal.IsTerminal(int(os.Stderr.Fd()))
}

// Get terminal dimensions
func setTerminalWidthVar() error {
	width, _, err := terminal.GetSize(int(os.Stderr.Fd()))
	// -5 to avoid edges
	terminalWidth = width - 5
	return err
}

// Initializes a new progress bar for headline, with a spinner
func (p *progressBarManager) newHeadlineBar(headline string) {
	p.barsWg.Add(1)
	p.headlineBar = p.container.AddSpinner(1, mpb.SpinnerOnLeft,
		mpb.SpinnerStyle([]string{"-", "-", "\\", "\\", "|", "|", "/", "/"}),
		mpb.BarRemoveOnComplete(),
		mpb.PrependDecorators(
			decor.Name(headline),
		),
	)
}

// Initializes a new progress bar that states the log file path. The bar's text remains after cli is done.
func (p *progressBarManager) printLogFilePathAsBar(path string) {
	p.barsWg.Add(1)
	prefix := " Log path: "
	p.logFilePathBar = p.container.AddBar(0,
		mpb.BarClearOnComplete(),
		mpb.PrependDecorators(
			decor.Name(buildDescByLimits(terminalWidth, prefix, path, "")),
		),
	)
	p.logFilePathBar.SetTotal(0, true)
}
