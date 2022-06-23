package log_test

import (
	"bytes"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/credifranco/stori-utils-go/log"
	"go.uber.org/zap"
)

// MemorySink implements zap.Sink by writing all messages to a buffer.
type MemorySink struct {
	*bytes.Buffer
}

// Implement Close and Sync as no-ops to satisfy the interface. The Write
// method is provided by the embedded buffer.
func (s *MemorySink) Close() error { return nil }
func (s *MemorySink) Sync() error  { return nil }

// Test that the logger config function creates configs with the correct log levels
func TestLoggerConfig(t *testing.T) {
	// Create a sink instance, and register it with zap for the "memory"
	// protocol.
	sink := &MemorySink{new(bytes.Buffer)}
	err := zap.RegisterSink("memory", func(*url.URL) (zap.Sink, error) {
		return sink, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create logger configs
	debugCfg := log.NewLoggerConfig(zap.NewAtomicLevelAt(zap.DebugLevel))
	infoCfg := log.NewLoggerConfig(zap.NewAtomicLevelAt(zap.InfoLevel))
	warnCfg := log.NewLoggerConfig(zap.NewAtomicLevelAt(zap.WarnLevel))
	errorCfg := log.NewLoggerConfig(zap.NewAtomicLevelAt(zap.ErrorLevel))

	// Redirect all messages to the MemorySink.
	infoCfg.OutputPaths = []string{"memory://"}
	debugCfg.OutputPaths = []string{"memory://"}
	warnCfg.OutputPaths = []string{"memory://"}
	errorCfg.OutputPaths = []string{"memory://"}

	dl, err := debugCfg.Build()
	if err != nil {
		t.Fatal(err)
	}
	il, err := infoCfg.Build()
	if err != nil {
		t.Fatal(err)
	}
	wl, err := warnCfg.Build()
	if err != nil {
		t.Fatal(err)
	}
	el, err := errorCfg.Build()
	if err != nil {
		t.Fatal(err)
	}

	debugLogger := dl.Sugar()
	infoLogger := il.Sugar()
	warnLogger := wl.Sugar()
	errorLogger := el.Sugar()

	// Test that the Debug level is correctly captured by the logger.
	debugLogger.Debugw("Debug test", "debug", "debug")
	output := sink.String()
	t.Logf("output = %s", output)
	if !strings.Contains(output, `"debug":"debug"`) {
		t.Error("Incorrect log output")
	}

	//Reset buffer
	sink.Reset()

	// Test that the Info level is correctly captured by the logger.
	infoLogger.Infow("Info test", "info", "info")
	output = sink.String()
	t.Logf("output = %s", output)
	if !strings.Contains(output, `"info":"info"`) {
		t.Error("Incorrect log output")
	}

	//Reset buffer
	sink.Reset()

	// Test that the Warn level is correctly captured by the logger.
	warnLogger.Warnw("Warn test", "warn", "warn")
	output = sink.String()
	t.Logf("output = %s", output)
	if !strings.Contains(output, `"warn":"warn"`) {
		t.Error("Incorrect log output")
	}

	//Reset buffer
	sink.Reset()

	// Test that the Error level is correctly captured by the logger.
	errorLogger.Errorw("Error test", "error", "error")
	output = sink.String()
	t.Logf("output = %s", output)
	if !strings.Contains(output, `"error":"error"`) {
		t.Error("Incorrect log output")
	}
}

// Test that the logger uses the correct log level based on the environment variable
func TestLogger(t *testing.T) {

	// Test when the LOG_LEVEL environment variable is not set
	l, err := log.NewLogger()
	if err != nil {
		t.Fatal(err)
	}
	if !l.Desugar().Core().Enabled(zap.InfoLevel) {
		t.Error("info logger should be enabled")
	}

	// Test when the LOG_LEVEL environment variable is set to DEBUG
	os.Setenv("LOG_LEVEL", "DEBUG")
	l, err = log.NewLogger()
	if err != nil {
		t.Fatal(err)
	}
	if !l.Desugar().Core().Enabled(zap.DebugLevel) {
		t.Error("debug logger should be enabled")
	}

	// Test when the LOG_LEVEL environment variable is set to INFO
	os.Setenv("LOG_LEVEL", "INFO")
	l, err = log.NewLogger()
	if err != nil {
		t.Fatal(err)
	}
	if !l.Desugar().Core().Enabled(zap.InfoLevel) {
		t.Error("info logger should be enabled")
	}

	// Test when the LOG_LEVEL environment variable is set to WARN
	os.Setenv("LOG_LEVEL", "WARN")
	l, err = log.NewLogger()
	if err != nil {
		t.Fatal(err)
	}
	if !l.Desugar().Core().Enabled(zap.WarnLevel) {
		t.Error("warn logger should be enabled")
	}

	// Test when the LOG_LEVEL environment variable is set to ERROR
	os.Setenv("LOG_LEVEL", "ERROR")
	l, err = log.NewLogger()
	if err != nil {
		t.Fatal(err)
	}
	if !l.Desugar().Core().Enabled(zap.ErrorLevel) {
		t.Error("error logger should be enabled")
	}

}
