package telemetry

import (
	"log/slog"
	"testing"
)

func TestIsRunningOnGCP(t *testing.T) {
	tests := []struct {
		name     string
		project  string
		kService string
		expected bool
	}{
		{"neither set", "", "", false},
		{"project set", "my-project", "", true},
		{"k_service set", "", "my-service", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("GOOGLE_CLOUD_PROJECT", test.project)
			t.Setenv("K_SERVICE", test.kService)
			if result := IsRunningOnGCP(); result != test.expected {
				t.Errorf("expected %v but got %v", test.expected, result)
			}
		})
	}
}

func TestReplaceAttrForGCP(t *testing.T) {
	tests := []struct {
		name     string
		groups   []string
		attr     slog.Attr
		expected string
	}{
		{"level renamed", nil, slog.Attr{Key: slog.LevelKey}, GCPSeverityKey},
		{"message renamed", nil, slog.Attr{Key: slog.MessageKey}, GCPMessageKey},
		{"time renamed", nil, slog.Attr{Key: slog.TimeKey}, GCPTimestampKey},
		{"other left alone", nil, slog.Attr{Key: "custom"}, "custom"},
		{"grouped attr left alone", []string{"group"}, slog.Attr{Key: slog.LevelKey}, slog.LevelKey},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ReplaceAttrForGCP(test.groups, test.attr)
			if result.Key != test.expected {
				t.Errorf("expected key %q but got %q", test.expected, result.Key)
			}
		})
	}
}

func TestNewGCPConsoleHandler(t *testing.T) {
	t.Run("not on GCP returns fallback", func(t *testing.T) {
		t.Setenv("GOOGLE_CLOUD_PROJECT", "")
		t.Setenv("K_SERVICE", "")
		fallback := slog.Default().Handler()
		result := NewGCPConsoleHandler(fallback)
		if result != fallback {
			t.Errorf("expected fallback handler to be returned")
		}
	})

	t.Run("on GCP returns JSON handler", func(t *testing.T) {
		t.Setenv("GOOGLE_CLOUD_PROJECT", "my-project")
		fallback := slog.Default().Handler()
		result := NewGCPConsoleHandler(fallback)
		if result == fallback {
			t.Errorf("expected a JSON handler, not the fallback")
		}
	})
}
