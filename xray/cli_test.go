package xray

import (
	"github.com/jfrog/jfrog-cli-core/v2/xray/commands/offlineupdate"
	"reflect"
	"testing"
)

func TestValidateStream(t *testing.T) {
	type args struct {
		streams string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"empty array", args{streams: ""}, "", true},
		{"PublicData", args{streams: offlineupdate.PublicData}, offlineupdate.PublicData, false},
		{"ContextualAnalysis", args{streams: offlineupdate.ContextualAnalysis}, offlineupdate.ContextualAnalysis, false},
		{"Exposures", args{streams: offlineupdate.Exposures}, offlineupdate.Exposures, false},
		{"invalid elements", args{streams: "bad_stream"}, "", true},
		{"array", args{streams: offlineupdate.PublicData + ";" + offlineupdate.ContextualAnalysis}, "", true},
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
