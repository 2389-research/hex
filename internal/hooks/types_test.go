package hooks

import "testing"

func TestHookConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  HookConfig
		wantErr bool
	}{
		{
			name: "valid hook with command",
			config: HookConfig{
				Command: "echo 'test'",
			},
			wantErr: false,
		},
		{
			name: "empty command",
			config: HookConfig{
				Command: "",
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			config: HookConfig{
				Command: "echo 'test'",
				Timeout: -5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
