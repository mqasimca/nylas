package auth

import (
	"bytes"
	"strings"
	"testing"
)

func TestProvidersCmd(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantOutput []string
		wantErr    bool
	}{
		{
			name:       "list providers",
			args:       []string{},
			wantOutput: []string{"Available Authentication Providers"},
			wantErr:    false,
		},
		{
			name:       "list providers json",
			args:       []string{"--json"},
			wantOutput: []string{},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newProvidersCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("newProvidersCmd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := buf.String()
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("newProvidersCmd() output = %v, want to contain %v", output, want)
				}
			}
		})
	}
}
