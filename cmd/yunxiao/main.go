package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aliyun/yunxiao-cli/internal/cli"
	commandauth "github.com/aliyun/yunxiao-cli/internal/command/auth"
	"github.com/aliyun/yunxiao-cli/internal/command/codeup"
	"github.com/aliyun/yunxiao-cli/internal/command/flow"
	"github.com/aliyun/yunxiao-cli/internal/command/meta"
	"github.com/aliyun/yunxiao-cli/internal/command/org"
	"github.com/aliyun/yunxiao-cli/internal/command/packages"
	"github.com/aliyun/yunxiao-cli/internal/command/projex"
	"github.com/aliyun/yunxiao-cli/internal/command/raw"
	"github.com/aliyun/yunxiao-cli/internal/command/testhub"
	"github.com/aliyun/yunxiao-cli/internal/config"
	"github.com/aliyun/yunxiao-cli/internal/model/output"
)

func main() {
	if err := config.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %s\n", err.Error())
		os.Exit(cli.WriteError(&output.ErrorDetail{
			Code:      "CONFIG_READ_FAILED",
			Category:  "general",
			Retryable: false,
			Message:   err.Error(),
		}, &output.Meta{}, cli.GetOutputFormat()))
	}

	root := cli.NewRootCmd()
	meta.InstallJSONHelp(root)
	root.AddCommand(meta.NewCommandsCmd(root))
	root.AddCommand(commandauth.NewAuthCmd())
	root.AddCommand(org.NewOrgCmd())
	root.AddCommand(codeup.NewCodeupCmd())
	root.AddCommand(flow.NewFlowCmd())
	root.AddCommand(projex.NewProjexCmd())
	root.AddCommand(packages.NewPackagesCmd())
	root.AddCommand(testhub.NewTesthubCmd())
	root.AddCommand(raw.NewRawCmd())

	if err := root.Execute(); err != nil {
		if code, ok := cli.ExitCode(err); ok {
			os.Exit(code)
		}
		errorCode := "COMMAND_FAILED"
		if strings.Contains(err.Error(), "invalid argument") && strings.Contains(err.Error(), " flag") {
			errorCode = "PARAM_INVALID"
		}
		fmt.Fprintf(os.Stderr, "[ERROR] command failed: %s\n", err.Error())
		os.Exit(cli.WriteError(&output.ErrorDetail{
			Code:      errorCode,
			Category:  "param",
			Retryable: false,
			Message:   err.Error(),
		}, &output.Meta{}, cli.GetOutputFormat()))
	}
}
