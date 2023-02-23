package xray

import (
	"github.com/jfrog/jfrog-cli-core/v2/xray/commands/offlineupdate"
	"reflect"
	"testing"
)

func TestValidateStream(t *testing.T) {
	streams := offlineupdate.NewValidStreams()
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
		{"PublicData", args{streams: streams.GetPublicDataStream()}, streams.GetPublicDataStream(), false},
		{"ContextualAnalysis", args{streams: streams.GetContextualAnalysisStream()}, streams.GetContextualAnalysisStream(), false},
		{"Exposures", args{streams: streams.GetExposuresStream()}, streams.GetExposuresStream(), false},
		{"invalid elements", args{streams: "bad_stream"}, "", true},
		{"array", args{streams: streams.GetPublicDataStream() + ";" + streams.GetContextualAnalysisStream()}, "", true},
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
