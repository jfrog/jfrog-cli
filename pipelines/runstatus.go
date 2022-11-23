package pipelines

import (
	"context"
	"fmt"
	"github.com/gen2brain/beeep"
	"github.com/gookit/color"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/pipelines"
	"github.com/jfrog/jfrog-client-go/pipelines/services"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
	"time"
)

var reStatus string

/*
 * fetchLatestPipelineRunStatus fetch pipeline run status based on flags
 * supplied
 */
func fetchLatestPipelineRunStatus(c *cli.Context, branch string) error {
	clientlog.Info(coreutils.PrintTitle("🐸🐸🐸 fetching pipeline run status"))

	serverID := c.String("server-id")
	pipName := c.String("name")
	notify := c.Bool("monitor")
	serviceDetails, err2 := getServiceDetails(serverID)
	if err2 != nil {
		return err2
	}

	pipelinesMgr, err3 := getPipelinesManager(serviceDetails)
	if err3 != nil {
		return err3
	}

	byBranch, err := pipelinesMgr.GetPipelineRunStatusByBranch(branch, pipName)
	if err != nil {
		return err
	}
	for i := range byBranch.Pipelines {
		p := byBranch.Pipelines[i]
		if p.LatestRunID != 0 {
			if pipName != "" && notify {
				err2 := monitorStatusAndNotify(context.Background(), pipelinesMgr, branch, pipName)
				if err2 != nil {
					return err2
				}
			} else {
				reStatus, colorCode, d := getPipelineStatusAndColorCode(&p)
				res := colorCode.Sprintf("\n%s %s\n%14s %s\n%14s %d \n%14s %s \n%14s %s %s\n", "PipelineName :", p.Name, "Branch :", p.PipelineSourceBranch, "Run :", p.Run.RunNumber, "Duration :", d, "Status :", reStatus)
				clientlog.Output(res)
			}
		}
	}
	return nil
}

/*
 * getPipelineStatusAndColorCode based on pipeline status code
 * return color to be used for pretty printing
 */
func getPipelineStatusAndColorCode(pipeline *services.Pipelines) (string, color.Color, string) {
	status := getPipelineStatus(pipeline.Run.StatusCode)
	colorCode := getStatusColorCode(status)
	d := pipeline.Run.DurationSeconds
	if d == 0 {
		t1 := pipeline.Run.StartedAt
		t2 := time.Now()
		d = int(t2.Sub(t1))
	}

	t := ConvertSecToDay(d)
	return status, colorCode, t
}

/*
 * check for change in status with the latest status
 */
func monitorStatusChange(pipStatus string) bool {
	if reStatus == pipStatus {
		return false
	}
	reStatus = pipStatus
	return true
}

/*
 * hasPipelineRunEnded if pipeline status is one of
 * CANCELLED, FAILED, SUCCESS, ERROR, TIMEOUT pipeline run
 * life is considered to be done.
 */
func hasPipelineRunEnded(pipStatus string) bool {
	pipRunEndLife := []string{SUCCESS, FAILURE, ERROR, CANCELLED, TIMEOUT}
	return contains(pipRunEndLife, pipStatus)
}

/*
 * monitorStatusAndNotify monitor for status change and
 * send notification if there is a change identified
 */
func monitorStatusAndNotify(ctx context.Context, pipelinesMgr *pipelines.PipelinesServicesManager, branch string, pipName string) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Minute)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			clientlog.Output("TIMED OUT!")
			return ctx.Err()
		default:
			p, err := pipelinesMgr.GetPipelineRunStatusByBranch(branch, pipName)
			if err != nil {
				return err
			}
			pipeline := p.Pipelines[0]
			status, colorCode, d := getPipelineDetails(pipeline)
			if monitorStatusChange(status) {
				res := colorCode.Sprintf("\n%s %s\n%14s %s\n%14s %d \n%14s %s \n%14s %s\n", "PipelineName :", pipeline.Name, "Branch :", pipeline.PipelineSourceBranch, "Run :", pipeline.Run.RunNumber, "Duration :", d, "Status :", reStatus)
				clientlog.Output(res)
				sendNotification(status, pipeline.Name)
				if hasPipelineRunEnded(status) {
					return nil
				}
			}
			reStatus = status
			time.Sleep(5 * time.Second)
		}
	}
}

/*
getPipelineDetails returns pipeline status, colorCode for matching status
and duration of pipeline
*/
func getPipelineDetails(pipeline services.Pipelines) (string, color.Color, string) {
	status := getPipelineStatus(pipeline.Run.StatusCode)
	colorCode := getStatusColorCode(status)
	d := pipeline.Run.DurationSeconds
	if d == 0 {
		t1 := pipeline.Run.StartedAt
		t2 := time.Now()
		d = int(t2.Sub(t1))
	}

	t := ConvertSecToDay(d)
	return status, colorCode, t
}

/*
 * getStatusColorCode returns gokit/color.Color
 * based on status input parameter
 */
func getStatusColorCode(status string) color.Color {
	colorCode := color.Blue
	if status == SUCCESS {
		return color.Green
	} else if status == FAILURE || status == ERROR || status == CANCELLED || status == TIMEOUT {
		return color.Red
	}
	return colorCode
}

/*
ConvertSecToDay converts seconds passed as integer to
 * D H M S format
*/
func ConvertSecToDay(sec int) string {
	day := sec / (24 * 3600)

	sec = sec % (24 * 3600)
	hour := sec / 3600

	sec %= 3600
	minutes := sec / 60

	sec %= 60
	seconds := sec

	v := fmt.Sprintf("%dD %dH %dM %dS", day, hour, minutes, seconds)
	return v
}

/*
sendNotification sends notification
*/
func sendNotification(pipStatus string, pipName string) {
	err := beeep.Alert("Pipelines CLI", pipName+" - "+pipStatus, "")
	if err != nil {
		panic(err)
	}
	yes := hasPipelineRunEnded(pipStatus)
	if yes {
		return
	}
}

/*
contains returns whether a string is available in slice
*/
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
