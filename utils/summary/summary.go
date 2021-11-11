package summary

import (
	"encoding/json"
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

func NewBuildInfoSummary(success, failed int, sha256 string, err error) *BuildInfoSummary {
	summaryReport := GetSummaryReport(success, failed, false, err)
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

type Sha256 struct {
	Sha256Str string `json:"sha256"`
}

func (bis *BuildInfoSummary) AddSha256(sha256Str string) {
	sha256 := Sha256{Sha256Str: sha256Str}
	bis.Sha256Array = append(bis.Sha256Array, sha256)
}

func GetSummaryReport(success, failed int, failNoOp bool, err error) *Summary {
	summary := &Summary{Totals: &Totals{}}
	if err != nil || failed > 0 {
		summary.Status = Failure
	} else if success == 0 && failNoOp {
		summary.Status = Failure
	} else {
		summary.Status = Success
	}
	summary.Totals.Success = success
	summary.Totals.Failure = failed
	return summary
}
