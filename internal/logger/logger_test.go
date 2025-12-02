package logger_test

import (
	"testing"

	"github.com/nais/kolide-check-validator/internal/logger"
)

func TestNew(t *testing.T) {
	t.Run("invalid log format", func(t *testing.T) {
		log, err := logger.New("foo", "info")
		if log != nil {
			t.Errorf("expected log to be nil, got %v", log)
		}

		if expected := `invalid log format: "foo"`; err.Error() != expected {
			t.Fatalf("expected error %q, got %q", expected, err.Error())
		}
	})

	t.Run("invalid log level", func(t *testing.T) {
		log, err := logger.New("json", "foo")
		if log != nil {
			t.Errorf("expected log to be nil, got %v", log)
		}

		if expected := `not a valid logrus Level: "foo"`; err.Error() != expected {
			t.Fatalf("expected error %q, got %q", expected, err.Error())
		}
	})
}
