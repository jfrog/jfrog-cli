package usage

import (
	"time"

	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	coreconfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

// DefaultReportTimeout bounds how long a command waits for the usage report to
// complete before giving up. Commands without a user-facing timeout flag (for
// example "jf mcp ...") should pass this so a slow report never blocks the CLI.
const DefaultReportTimeout = 2 * time.Second

// ReportUsageFn is the usage reporter invoked by StartReport. It is a package
// variable so tests can inject fakes (panic, slow, fast) without touching HTTP.
var ReportUsageFn = commands.ReportUsage

// StartReport collects command metrics and launches the usage report in a
// goroutine, recovering from any panic so that a misbehaving reporter cannot
// crash the process. The returned channel is signaled by ReportUsage on
// success, or closed by the recover path on panic — either way the caller's
// receive unblocks. Pair it with a deferred WaitForReport.
func StartReport(commandName string, flagsUsed []string, serverDetails *coreconfig.ServerDetails) chan bool {
	commands.CollectMetrics(commandName, flagsUsed)
	waitUsageReport := make(chan bool)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Debug(commandName+": usage report panicked:", r)
				close(waitUsageReport)
			}
		}()
		ReportUsageFn(commandName, serverDetails, waitUsageReport)
	}()
	return waitUsageReport
}

// WaitForReport blocks until the usage report completes or the timeout elapses.
// A non-positive timeout waits indefinitely.
func WaitForReport(commandName string, wait <-chan bool, timeout time.Duration) {
	if timeout <= 0 {
		<-wait
		return
	}
	select {
	case <-wait:
	case <-time.After(timeout):
		log.Debug(commandName+": usage report did not complete within", timeout)
	}
}
