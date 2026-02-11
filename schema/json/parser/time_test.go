package parser_test

import (
	"mokapi/schema/json/parser"
	"testing"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		wantErr bool
	}{
		{name: "invalid-duration-1", args: "T0S", wantErr: true},
		{name: "invalid-duration-2", args: "P-T0S", wantErr: true},
		{name: "invalid-duration-3", args: "PT1SP0D", wantErr: true},
		{name: "invalid-duration-4", args: "AT1SP0D", wantErr: true},
		{name: "invalid-duration-5", args: "P", wantErr: true},
		{name: "invalid-duration-6", args: "T", wantErr: true},
		{name: "invalid-duration-7", args: "-P", wantErr: true},
		{name: "invalid-unit-miss1", args: "P7Y4", wantErr: true},
		{name: "invalid-unit-miss2", args: "P6", wantErr: true},
		{name: "valid-period-only", args: "P2Y", wantErr: false},
		{name: "valid-time-decimal", args: "PT4.5S", wantErr: false},
		{name: "valid-full-iso8601", args: "P4Y3M2DT12H30M5.5S", wantErr: false},
		{name: "valid-seconds-1", args: "PT10S", wantErr: false},
		{name: "valid-minutes-2", args: "PT25M", wantErr: false},
		{name: "valid-hours-3", args: "PT2H", wantErr: false},
		{name: "valid-negative", args: "-PT5M", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.ParseDuration(tt.args)
			if tt.wantErr == true && err == nil {
				t.Errorf("ParseDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr == false && err != nil {
				t.Errorf("ParseDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
