package summary

import (
	"encoding/json"
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

func New(err error) *Summary {
	summary := &Summary{Totals: &Totals{}}
	if err != nil {
		summary.Status = Failure
	} else {
		summary.Status = Success
	}
	return summary
}

func (summary *Summary) Marshal() ([]byte, error) {
	return json.Marshal(summary)
}

type Summary struct {
	Status StatusType `json:"status"`
	Totals *Totals    `json:"totals"`
}

type Totals struct {
	Success int `json:"success"`
	Failure int `json:"failure"`
}
