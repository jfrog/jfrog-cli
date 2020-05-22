package utils

import "github.com/jfrog/jfrog-client-go/utils/io/httputils/responsereaderwriter"

type Result struct {
	successCount  int
	failCount     int
	resultsReader *responsereaderwriter.ResponseReader
}

func (r *Result) SuccessCount() int {
	return r.successCount
}

func (r *Result) FailCount() int {
	return r.failCount
}

func (r *Result) ResultsReader() *responsereaderwriter.ResponseReader {
	return r.resultsReader
}

func (r *Result) SetSuccessCount(successCount int) {
	r.successCount = successCount
}

func (r *Result) SetFailCount(failCount int) {
	r.failCount = failCount
}

func (r *Result) SetResultsReader(reader *responsereaderwriter.ResponseReader) {
	r.resultsReader = reader
}
