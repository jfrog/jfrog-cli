package xray

import (
	"github.com/jfrog/jfrog-cli-core/v2/xray/commands/offlineupdate"
	"reflect"
	"testing"
)

func TestValidateStream(t *testing.T) {
	type args struct {
		streams []string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]bool
		wantErr bool
	}{
		{"array nil", args{streams: nil}, map[string]bool{}, false},
		{"empty array", args{streams: []string{}}, map[string]bool{}, false},
		{"one element in array", args{streams: []string{offlineupdate.ContextualAnalysis}}, map[string]bool{offlineupdate.ContextualAnalysis: true}, false},
		{"two element in array", args{streams: []string{offlineupdate.PublicData, offlineupdate.Exposures}}, map[string]bool{offlineupdate.Exposures: true, offlineupdate.PublicData: true}, false},
		{"duplication", args{streams: []string{offlineupdate.PublicData, offlineupdate.ContextualAnalysis, offlineupdate.Exposures, offlineupdate.PublicData, offlineupdate.ContextualAnalysis, offlineupdate.Exposures}}, map[string]bool{offlineupdate.PublicData: true, offlineupdate.ContextualAnalysis: true, offlineupdate.Exposures: true}, false},
		{"invalid elements", args{streams: []string{"bad", "element"}}, nil, true},
		{"valid and invalid elements", args{streams: []string{offlineupdate.PublicData, "bad", "element"}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateStream(tt.args.streams)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("validateStream() got = %v, want %v", got, tt.want)
			}
		})
	}
}
