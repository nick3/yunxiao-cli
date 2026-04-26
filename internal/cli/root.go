package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	globalJSON    bool
	globalHuman   bool
	globalQuiet   bool
	globalVerbose bool
	globalDebug   bool
	globalTimeout int
	globalNoRetry bool
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "yunxiao",
		Short:         "Yunxiao CLI - Agent-first DevOps command line tool",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().BoolVar(&globalJSON, "json", false, "Force JSON output")
	root.PersistentFlags().BoolVar(&globalHuman, "human", false, "Force human-readable output")
	root.PersistentFlags().BoolVar(&globalQuiet, "quiet", false, "Suppress all stderr except errors")
	root.PersistentFlags().BoolVar(&globalVerbose, "verbose", false, "Show info-level diagnostics on stderr")
	root.PersistentFlags().BoolVar(&globalDebug, "debug", false, "Show all diagnostics including HTTP details on stderr")
	root.PersistentFlags().IntVar(&globalTimeout, "timeout", 30, "Request timeout in seconds")
	root.PersistentFlags().BoolVar(&globalNoRetry, "no-retry", false, "Disable built-in retry for transient errors")
	root.PersistentFlags().String("organization-id", "", "Organization ID")
	root.PersistentFlags().String("region", "", "Region selector")
	root.PersistentFlags().String("trace-id", "", "Trace ID for request correlation")

	return root
}

func GetOutputFormat() OutputFormat {
	if globalJSON {
		return FormatJSON
	}
	if globalHuman {
		return FormatHuman
	}
	return FormatAuto
}

func Execute() int {
	root := NewRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ExitGeneralFailure
	}
	return ExitSuccess
}
