package broker

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
		textErr string
	}{
		{
			"Undefined QueueError",
			&Config{},
			true,
			"error queue name is undefined",
		},
		{
			"Undefined QueueInvalid",
			&Config{QueueError: "error"},
			true,
			"invalid queue name is undefined",
		},
		{
			"Undefined QueueMain",
			&Config{QueueError: "error", QueueInvalid: "invalid"},
			true,
			"main queue name is undefined",
		},
		{
			"Undefined RepositoryConfig",
			&Config{QueueError: "error", QueueInvalid: "invalid", QueueMain: "main"},
			true,
			"repository config is undefined",
		},
		{
			"Undefined DynamoDBTable",
			&Config{QueueError: "error", QueueInvalid: "invalid", QueueMain: "main", RepositoryConfig: &RepositoryConfig{Backend: "dynamodb"}},
			true,
			"repository config is missing details",
		},
		{
			"Valid",
			&Config{QueueError: "error", QueueInvalid: "invalid", QueueMain: "main", RepositoryConfig: &RepositoryConfig{Backend: "builtin"}},
			false,
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error not received")
				}
				if err.Error() != tt.textErr {
					t.Errorf("unexpected error text; got %s, want %s", err.Error(), tt.textErr)
				}
			}
		})
	}
}

func TestConfig_SetValidationMode(t *testing.T) {
	testCases := []struct {
		input string
		want  ValidationMode
	}{
		{"false", ValidationModeDisabled},
		{"warnings", ValidationModeWarnings},
		{"true", ValidationModeStrict},
		{"", ValidationModeStrict},
	}
	config := &Config{}
	for _, tc := range testCases {
		config.SetValidationMode(tc.input)
		if have := config.Validation; have != tc.want {
			t.Errorf("SetValidationMode(); given %s, have %v, want %v", tc.input, have, tc.want)
		}
	}
}
