package meta

import (
	"fmt"
	"os"
	"strings"

	"github.com/aliyun/yunxiao-cli/internal/cli"
	"github.com/aliyun/yunxiao-cli/internal/command/flagmeta"
	"github.com/aliyun/yunxiao-cli/internal/model/output"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type CommandSpec struct {
	Path        string        `json:"path"`
	Short       string        `json:"short,omitempty"`
	Flags       []FlagSpec    `json:"flags,omitempty"`
	Subcommands []CommandSpec `json:"subcommands,omitempty"`
}

type FlagSpec struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Usage    string `json:"usage,omitempty"`
}

var writeResult = cli.WriteResult

func NewCommandsCmd(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "commands",
		Short: "List commands in structured JSON format",
		RunE: func(cmd *cobra.Command, args []string) error {
			specs := make([]CommandSpec, 0, len(root.Commands()))
			for _, sub := range root.Commands() {
				if !sub.IsAvailableCommand() || sub.Name() == "help" || sub.Name() == "completion" || sub.Name() == "commands" {
					continue
				}
				specs = append(specs, BuildSpec(sub, root.Name()))
			}
			if code := writeResult(specs, &output.Meta{}, cli.GetOutputFormat()); code != cli.ExitSuccess {
				return cli.NewExitError(code, "failed to write commands output")
			}
			return nil
		},
	}
}

func InstallJSONHelp(root *cobra.Command) {
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			if code := writeResult(BuildSpec(cmd, root.Name()), &output.Meta{}, cli.GetOutputFormat()); code != cli.ExitSuccess {
				os.Exit(code)
			}
			return
		}
		if err := cmd.Usage(); err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] failed to write help usage: %v\n", err)
		}
	})
}

func BuildSpec(cmd *cobra.Command, prefix string) CommandSpec {
	path := commandPath(cmd, prefix)
	spec := CommandSpec{Path: path, Short: cmd.Short, Flags: collectFlags(cmd)}
	for _, sub := range cmd.Commands() {
		if !sub.IsAvailableCommand() || sub.Name() == "help" || sub.Name() == "completion" {
			continue
		}
		spec.Subcommands = append(spec.Subcommands, BuildSpec(sub, path))
	}
	return spec
}

func commandPath(cmd *cobra.Command, fallback string) string {
	if cmd.CommandPath() != "" {
		return cmd.CommandPath()
	}
	return strings.TrimSpace(fallback + " " + cmd.Name())
}

func collectFlags(cmd *cobra.Command) []FlagSpec {
	flags := make([]FlagSpec, 0)
	cmd.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
		flags = append(flags, flagSpec(flag))
	})
	cmd.NonInheritedFlags().VisitAll(func(flag *pflag.Flag) {
		if flag.Name == "help" {
			return
		}
		flags = append(flags, flagSpec(flag))
	})
	return flags
}

func flagSpec(flag *pflag.Flag) FlagSpec {
	_, required := flag.Annotations[flagmeta.RequiredFlagAnnotation]
	return FlagSpec{Name: flag.Name, Type: flag.Value.Type(), Required: required, Usage: flag.Usage}
}
