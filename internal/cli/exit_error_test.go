package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExitCodeExtractsWrappedExitError(t *testing.T) {
	err := errors.Join(NewExitError(ExitGeneralFailure, "write failed"), errors.New("context"))

	code, ok := ExitCode(err)

	require.True(t, ok)
	require.Equal(t, ExitGeneralFailure, code)
}
