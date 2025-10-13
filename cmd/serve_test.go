package main

import (
	"testing"
)

func TestServeCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "help flag",
			args:    []string{"serve", "--help"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := rootCmd
			cmd.SetArgs(tt.args)
			err := executeCommand(cmd)

			if (err != nil) != tt.wantErr {
				t.Errorf("serve command error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

