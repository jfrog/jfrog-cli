package usage

import (
	"testing"
	"time"

	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	coreconfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// swapReportUsageFn replaces ReportUsageFn for the duration of the test and
// restores it on cleanup. Tests using this must not run in parallel.
func swapReportUsageFn(t *testing.T, fn func(string, *coreconfig.ServerDetails, chan<- bool)) {
	t.Helper()
	prev := ReportUsageFn
	ReportUsageFn = fn
	t.Cleanup(func() { ReportUsageFn = prev })
}

func TestStartReport_CollectsMetrics(t *testing.T) {
	swapReportUsageFn(t, func(_ string, _ *coreconfig.ServerDetails, ch chan<- bool) {
		ch <- true
	})

	flags := []string{"verbose", "header"}
	wait := StartReport("jf usage-test", flags, &coreconfig.ServerDetails{})
	WaitForReport("jf usage-test", wait, time.Second)

	got := commands.GetCollectedMetrics("jf usage-test")
	require.NotNil(t, got, "expected metrics to be collected for command \"jf usage-test\"")
	assert.Equal(t, flags, got.Flags)
}

func TestStartReport_PanicIsRecovered(t *testing.T) {
	swapReportUsageFn(t, func(_ string, _ *coreconfig.ServerDetails, _ chan<- bool) {
		panic("boom")
	})

	wait := StartReport("jf usage-test", nil, &coreconfig.ServerDetails{})
	select {
	case <-wait:
		// Channel was closed by the recover path; receive returns the zero value.
	case <-time.After(time.Second):
		t.Fatal("StartReport did not unblock after panic")
	}
}

func TestWaitForReport(t *testing.T) {
	const cmd = "jf usage-test"

	t.Run("non-positive timeout waits for signal", func(t *testing.T) {
		ch := make(chan bool, 1)
		ch <- true
		start := time.Now()
		WaitForReport(cmd, ch, 0)
		assert.Less(t, time.Since(start), 50*time.Millisecond)
	})

	t.Run("returns immediately when signaled before timeout", func(t *testing.T) {
		ch := make(chan bool, 1)
		ch <- true
		start := time.Now()
		WaitForReport(cmd, ch, time.Second)
		assert.Less(t, time.Since(start), 50*time.Millisecond)
	})

	t.Run("returns at timeout when never signaled", func(t *testing.T) {
		ch := make(chan bool)
		start := time.Now()
		WaitForReport(cmd, ch, 100*time.Millisecond)
		elapsed := time.Since(start)
		assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond)
		assert.Less(t, elapsed, 500*time.Millisecond)
	})
}
