package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/aliyun/yunxiao-cli/internal/model/output"
	"golang.org/x/term"
)

var (
	stdoutWriter io.Writer = os.Stdout
	stderrWriter io.Writer = os.Stderr
)

type OutputFormat int

const (
	FormatAuto OutputFormat = iota
	FormatJSON
	FormatHuman
)

func WriteResult(data interface{}, meta *output.Meta, format OutputFormat) int {
	env := &output.Envelope{
		Version: "v1",
		Data:    data,
		Meta:    meta,
		Error:   nil,
	}
	if shouldOutputJSON(format) {
		if err := writeJSON(env); err != nil {
			fmt.Fprintf(stderrWriter, "[ERROR] failed to write JSON envelope: %v\n", err)
			return ExitGeneralFailure
		}
		return ExitSuccess
	}
	if err := writeHuman(data); err != nil {
		fmt.Fprintf(stderrWriter, "[ERROR] failed to write human output: %v\n", err)
		return ExitGeneralFailure
	}
	return ExitSuccess
}

func WriteError(err *output.ErrorDetail, meta *output.Meta, format OutputFormat) int {
	env := &output.Envelope{
		Version: "v1",
		Data:    nil,
		Meta:    meta,
		Error:   err,
	}
	if writeErr := writeJSON(env); writeErr != nil {
		fmt.Fprintf(stderrWriter, "[ERROR] failed to write JSON envelope: %v\n", writeErr)
		return ExitGeneralFailure
	}
	return exitCodeFromCategory(err.Category)
}

func shouldOutputJSON(format OutputFormat) bool {
	if format == FormatJSON {
		return true
	}
	if format == FormatHuman {
		return false
	}
	return !term.IsTerminal(int(os.Stdout.Fd()))
}

func writeJSON(env *output.Envelope) error {
	enc := json.NewEncoder(stdoutWriter)
	enc.SetIndent("", "  ")
	return enc.Encode(env)
}

func writeHuman(data interface{}) error {
	if data == nil {
		return nil
	}
	_, err := fmt.Fprintf(stdoutWriter, "%v\n", data)
	return err
}

func exitCodeFromCategory(category string) int {
	switch category {
	case "auth":
		return ExitAuthFailed
	case "forbidden":
		return ExitForbidden
	case "not_found":
		return ExitNotFound
	case "rate_limit":
		return ExitRateLimited
	case "network":
		return ExitNetworkFailure
	case "upstream":
		return ExitUpstreamError
	case "param":
		return ExitParamError
	default:
		return ExitGeneralFailure
	}
}
