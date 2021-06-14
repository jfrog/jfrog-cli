package summary

import (
	"encoding/json"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"strings"
)

type StatusType int

const (
	Success StatusType = iota
	Failure
)

var StatusTypes = []string{
	"success",
	"failure",
}

func (statusType StatusType) MarshalJSON() ([]byte, error) {
	return json.Marshal(StatusTypes[statusType])
}

func (statusType *StatusType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	default:
		*statusType = Failure
	case "success":
		*statusType = Success

	}
	return nil
}

func NewSummary(err error) *Summary {
	summary := &Summary{Totals: &Totals{}}
	if err != nil {
		summary.Status = Failure
	} else {
		summary.Status = Success
	}
	return summary
}

func NewBuildInfoSummary(success, failed int, sha256 string, err error) *BuildInfoSummary {
	summaryReport := GetSummaryReport(success, failed, err)
	buildInfoSummary := BuildInfoSummary{Summary: *summaryReport, Sha256Array: []Sha256{}}
	if success == 1 {
		buildInfoSummary.AddSha256(sha256)
	}
	return &buildInfoSummary
}

func (summary *Summary) Marshal() ([]byte, error) {
	return json.Marshal(summary)
}

func (bis *BuildInfoSummary) Marshal() ([]byte, error) {
	return json.Marshal(bis)
}

type Summary struct {
	Status StatusType `json:"status"`
	Totals *Totals    `json:"totals"`
}

type Totals struct {
	Success int `json:"success"`
	Failure int `json:"failure"`
}

type BuildInfoSummary struct {
	Summary
	Sha256Array []Sha256 `json:"files"`
}

type GoPublishSummary struct {
	Summary
	Files []clientutils.FileTransferDetails `json:"files"`
}

type Sha256 struct {
	Sha256Str string `json:"sha256"`
}

func (bis *BuildInfoSummary) AddSha256(sha256Str string) {
	sha256 := Sha256{Sha256Str: sha256Str}
	bis.Sha256Array = append(bis.Sha256Array, sha256)
}

func GetSummaryReport(success, failed int, err error) *Summary {
	summaryReport := NewSummary(err)
	summaryReport.Totals.Success = success
	summaryReport.Totals.Failure = failed
	if err == nil && summaryReport.Totals.Failure != 0 {
		summaryReport.Status = Failure
	}
	return summaryReport
}
