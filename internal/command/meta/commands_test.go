package meta

import (
	"testing"

	"github.com/nick3/yunxiao-cli/internal/cli"
	"github.com/nick3/yunxiao-cli/internal/model/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestCommandsCmdReturnsExitErrorWhenOutputWriteFails(t *testing.T) {
	oldWriteResult := writeResult
	writeResult = func(data any, meta *output.Meta, format cli.OutputFormat) int {
		return cli.ExitGeneralFailure
	}
	t.Cleanup(func() { writeResult = oldWriteResult })

	root := &cobra.Command{Use: "yunxiao"}
	root.AddCommand(&cobra.Command{Use: "org", Short: "Organization commands"})
	cmd := NewCommandsCmd(root)

	err := cmd.RunE(cmd, nil)

	require.Error(t, err)
	code, ok := cli.ExitCode(err)
	require.True(t, ok)
	require.Equal(t, cli.ExitGeneralFailure, code)
}

func TestBuildSpecDoesNotMarkOptionalFlagRequiredByNameAlone(t *testing.T) {
	cmd := &cobra.Command{Use: "example"}
	cmd.Flags().String("project-id", "", "Optional project ID")

	spec := BuildSpec(cmd, "yunxiao")

	require.Len(t, spec.Flags, 1)
	require.Equal(t, "project-id", spec.Flags[0].Name)
	require.False(t, spec.Flags[0].Required)
}
