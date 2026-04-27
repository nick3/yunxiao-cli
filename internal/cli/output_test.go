package cli

import (
	"errors"
	"io"
	"testing"

	"github.com/nick3/yunxiao-cli/internal/model/output"
	"github.com/stretchr/testify/require"
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestWriteResultReturnsGeneralFailureWhenJSONWriteFails(t *testing.T) {
	oldStdout := stdoutWriter
	oldStderr := stderrWriter
	stdoutWriter = failingWriter{}
	stderrWriter = io.Discard
	t.Cleanup(func() {
		stdoutWriter = oldStdout
		stderrWriter = oldStderr
	})

	code := WriteResult(map[string]any{"ok": true}, &output.Meta{}, FormatJSON)

	require.Equal(t, ExitGeneralFailure, code)
}

func TestWriteErrorReturnsGeneralFailureWhenJSONWriteFails(t *testing.T) {
	oldStdout := stdoutWriter
	oldStderr := stderrWriter
	stdoutWriter = failingWriter{}
	stderrWriter = io.Discard
	t.Cleanup(func() {
		stdoutWriter = oldStdout
		stderrWriter = oldStderr
	})

	code := WriteError(&output.ErrorDetail{Code: "AUTH_FAILED", Category: "auth", Message: "auth failed"}, &output.Meta{}, FormatJSON)

	require.Equal(t, ExitGeneralFailure, code)
}

func TestCommandResultWriteFailureReturnsGeneralFailure(t *testing.T) {
	oldStdout := stdoutWriter
	oldStderr := stderrWriter
	stdoutWriter = failingWriter{}
	stderrWriter = io.Discard
	t.Cleanup(func() {
		stdoutWriter = oldStdout
		stderrWriter = oldStderr
	})

	code := WriteResult([]map[string]any{{"path": "yunxiao org current"}}, &output.Meta{}, FormatJSON)

	require.Equal(t, ExitGeneralFailure, code)
}
