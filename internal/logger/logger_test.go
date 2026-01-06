package logger

import "testing"

func TestLogHookPanicDoesNotCrash(t *testing.T) {
	l := GetLogger()
	l.SetHook(func(levelStr string, message string) {
		panic("boom")
	})

	// If the hook panics and isn't recovered, the test process would crash.
	l.Error("test error")
}
