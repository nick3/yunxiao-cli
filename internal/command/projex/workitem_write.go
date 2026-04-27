package projex

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/nick3/yunxiao-cli/internal/cli"
	"github.com/nick3/yunxiao-cli/internal/command/flagmeta"
	"github.com/nick3/yunxiao-cli/internal/config"
	orgdomain "github.com/nick3/yunxiao-cli/internal/domains/org"
	projexdomain "github.com/nick3/yunxiao-cli/internal/domains/projex"
	"github.com/nick3/yunxiao-cli/internal/httpx"
	"github.com/nick3/yunxiao-cli/internal/model/output"
	"github.com/spf13/cobra"
)

const maxWorkitemTextFileSize = 1024 * 1024

type workitemWriteFlags struct {
	OrganizationID   string
	WorkitemID       string
	ProjectID        string
	SpaceID          string
	Subject          string
	WorkitemTypeID   string
	AssignedTo       string
	Description      string
	DescriptionFile  string
	FormatType       string
	Status           string
	Priority         string
	Labels           string
	ParentID         string
	Participants     string
	Sprint           string
	Trackers         string
	Verifier         string
	Versions         string
	CustomFields     []string
	CustomFieldsJSON string
	Yes              bool
}

func newWorkitemCreateCmd() *cobra.Command {
	var flags workitemWriteFlags
	cmd := &cobra.Command{Use: "create", Short: "Create a workitem", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID := config.GetOrganizationID(flags.OrganizationID)
		if !flags.Yes {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "yes is required because creating a workitem writes to Yunxiao")
			return nil
		}
		if orgID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "organization_id is required")
			return nil
		}
		spaceID, ok := resolveProjexProjectID(flags.ProjectID, flags.SpaceID, format, meta)
		if !ok {
			return nil
		}
		if spaceID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "project_id or space_id is required")
			return nil
		}
		if flags.Subject == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "subject is required")
			return nil
		}
		if flags.WorkitemTypeID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "workitem_type_id is required")
			return nil
		}
		if flags.AssignedTo == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "assigned_to is required")
			return nil
		}
		input, ok := buildWorkitemCreateInput(cmd, flags, spaceID, format, meta)
		if !ok {
			return nil
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		if !resolveAssignedToSelf(context.Background(), client, &input.AssignedTo, format, meta) {
			return nil
		}
		data, errDetail := projexdomain.CreateWorkitem(context.Background(), client, orgID, input)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] workitem create failed: %s\n", errDetail.Message)
			os.Exit(cli.WriteError(errDetail, meta, format))
			return nil
		}
		if code := cli.WriteResult(data, meta, format); code != cli.ExitSuccess {
			os.Exit(code)
		}
		return nil
	}}
	addWorkitemWriteFlags(cmd, &flags, true)
	flagmeta.MustMarkRequired(cmd, "organization-id", "subject", "workitem-type-id", "assigned-to", "yes")
	return cmd
}

func newWorkitemUpdateCmd() *cobra.Command {
	var flags workitemWriteFlags
	cmd := &cobra.Command{Use: "update", Short: "Update a workitem", RunE: func(cmd *cobra.Command, args []string) error {
		format := cli.GetOutputFormat()
		meta := &output.Meta{}
		orgID := config.GetOrganizationID(flags.OrganizationID)
		if !flags.Yes {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "yes is required because updating a workitem writes to Yunxiao")
			return nil
		}
		if orgID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "organization_id is required")
			return nil
		}
		if flags.WorkitemID == "" {
			exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "workitem_id is required")
			return nil
		}
		input, ok := buildWorkitemUpdateInput(cmd, flags, format, meta)
		if !ok {
			return nil
		}
		client, ok := newAPIClient(cmd, format, meta)
		if !ok {
			return nil
		}
		if !resolveAssignedToSelf(context.Background(), client, &input.AssignedTo, format, meta) {
			return nil
		}
		data, errDetail := projexdomain.UpdateWorkitem(context.Background(), client, orgID, flags.WorkitemID, input)
		if errDetail != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] workitem update failed: %s\n", errDetail.Message)
			os.Exit(cli.WriteError(errDetail, meta, format))
			return nil
		}
		if code := cli.WriteResult(data, meta, format); code != cli.ExitSuccess {
			os.Exit(code)
		}
		return nil
	}}
	addWorkitemWriteFlags(cmd, &flags, false)
	flagmeta.MustMarkRequired(cmd, "organization-id", "workitem-id", "yes")
	return cmd
}

func addWorkitemWriteFlags(cmd *cobra.Command, flags *workitemWriteFlags, create bool) {
	cmd.Flags().StringVar(&flags.OrganizationID, "organization-id", "", "Organization ID")
	if create {
		cmd.Flags().StringVar(&flags.ProjectID, "project-id", "", "Project ID; alias of space-id")
		cmd.Flags().StringVar(&flags.SpaceID, "space-id", "", "Projex space ID; alias of project-id")
		cmd.Flags().StringVar(&flags.WorkitemTypeID, "workitem-type-id", "", "Workitem type ID")
	} else {
		cmd.Flags().StringVar(&flags.WorkitemID, "workitem-id", "", "Workitem ID")
		cmd.Flags().StringVar(&flags.Status, "status", "", "Status ID")
		cmd.Flags().StringVar(&flags.Priority, "priority", "", "Priority ID")
	}
	cmd.Flags().StringVar(&flags.Subject, "subject", "", "Workitem subject")
	cmd.Flags().StringVar(&flags.AssignedTo, "assigned-to", "", "Assignee user ID, or self for current user")
	cmd.Flags().StringVar(&flags.Description, "description", "", "Workitem description")
	cmd.Flags().StringVar(&flags.DescriptionFile, "description-file", "", "Read workitem description from a UTF-8 text file")
	cmd.Flags().StringVar(&flags.FormatType, "format-type", "", "Description format type: RICHTEXT or MARKDOWN")
	cmd.Flags().StringVar(&flags.Labels, "labels", "", "Comma-separated label IDs")
	cmd.Flags().StringVar(&flags.Sprint, "sprint", "", "Sprint ID")
	cmd.Flags().StringVar(&flags.Trackers, "trackers", "", "Comma-separated tracker user IDs")
	cmd.Flags().StringVar(&flags.Verifier, "verifier", "", "Verifier user ID")
	cmd.Flags().StringVar(&flags.Participants, "participants", "", "Comma-separated participant user IDs")
	cmd.Flags().StringVar(&flags.Versions, "versions", "", "Comma-separated version IDs")
	cmd.Flags().StringArrayVar(&flags.CustomFields, "custom-field", nil, "Custom field as key=value; repeatable")
	cmd.Flags().StringVar(&flags.CustomFieldsJSON, "custom-fields-json", "", "Custom fields JSON object")
	cmd.Flags().BoolVar(&flags.Yes, "yes", false, "Confirm this write operation")
	if create {
		cmd.Flags().StringVar(&flags.ParentID, "parent-id", "", "Parent workitem ID")
	}
}

func buildWorkitemCreateInput(cmd *cobra.Command, flags workitemWriteFlags, spaceID string, format cli.OutputFormat, meta *output.Meta) (projexdomain.WorkitemCreateInput, bool) {
	description, ok := resolveTextValue(cmd, "description", flags.Description, flags.DescriptionFile, "description", format, meta)
	if !ok {
		return projexdomain.WorkitemCreateInput{}, false
	}
	if !validateFormatType(flags.FormatType, format, meta) {
		return projexdomain.WorkitemCreateInput{}, false
	}
	customFields, ok := mergeCustomFields(flags.CustomFields, flags.CustomFieldsJSON, format, meta)
	if !ok {
		return projexdomain.WorkitemCreateInput{}, false
	}
	labels, ok := csvFlag(cmd, "labels", flags.Labels, format, meta)
	if !ok {
		return projexdomain.WorkitemCreateInput{}, false
	}
	participants, ok := csvFlag(cmd, "participants", flags.Participants, format, meta)
	if !ok {
		return projexdomain.WorkitemCreateInput{}, false
	}
	trackers, ok := csvFlag(cmd, "trackers", flags.Trackers, format, meta)
	if !ok {
		return projexdomain.WorkitemCreateInput{}, false
	}
	versions, ok := csvFlag(cmd, "versions", flags.Versions, format, meta)
	if !ok {
		return projexdomain.WorkitemCreateInput{}, false
	}
	return projexdomain.WorkitemCreateInput{AssignedTo: flags.AssignedTo, SpaceID: spaceID, Subject: flags.Subject, WorkitemTypeID: flags.WorkitemTypeID, CustomFieldValues: customFields, Description: description, FormatType: flags.FormatType, Labels: labels, ParentID: flags.ParentID, Participants: participants, Sprint: flags.Sprint, Trackers: trackers, Verifier: flags.Verifier, Versions: versions}, true
}

func buildWorkitemUpdateInput(cmd *cobra.Command, flags workitemWriteFlags, format cli.OutputFormat, meta *output.Meta) (projexdomain.WorkitemUpdateInput, bool) {
	description, ok := resolveTextValue(cmd, "description", flags.Description, flags.DescriptionFile, "description", format, meta)
	if !ok {
		return projexdomain.WorkitemUpdateInput{}, false
	}
	if !validateFormatType(flags.FormatType, format, meta) {
		return projexdomain.WorkitemUpdateInput{}, false
	}
	customFields, ok := mergeCustomFields(flags.CustomFields, flags.CustomFieldsJSON, format, meta)
	if !ok {
		return projexdomain.WorkitemUpdateInput{}, false
	}
	labels, ok := csvFlag(cmd, "labels", flags.Labels, format, meta)
	if !ok {
		return projexdomain.WorkitemUpdateInput{}, false
	}
	trackers, ok := csvFlag(cmd, "trackers", flags.Trackers, format, meta)
	if !ok {
		return projexdomain.WorkitemUpdateInput{}, false
	}
	participants, ok := csvFlag(cmd, "participants", flags.Participants, format, meta)
	if !ok {
		return projexdomain.WorkitemUpdateInput{}, false
	}
	versions, ok := csvFlag(cmd, "versions", flags.Versions, format, meta)
	if !ok {
		return projexdomain.WorkitemUpdateInput{}, false
	}
	input := projexdomain.WorkitemUpdateInput{Subject: flags.Subject, Description: description, FormatType: flags.FormatType, Status: flags.Status, AssignedTo: flags.AssignedTo, Priority: flags.Priority, Labels: labels, Sprint: flags.Sprint, Trackers: trackers, Verifier: flags.Verifier, Participants: participants, Versions: versions, CustomFieldValues: customFields}
	if !hasWorkitemUpdate(input) {
		exitWithError(format, meta, "PARAM_REQUIRED", "param", false, "at least one update field is required")
		return projexdomain.WorkitemUpdateInput{}, false
	}
	return input, true
}

func hasWorkitemUpdate(input projexdomain.WorkitemUpdateInput) bool {
	return input.Subject != "" || input.Description != "" || input.FormatType != "" || input.Status != "" || input.AssignedTo != "" || input.Priority != "" || len(input.Labels) > 0 || input.Sprint != "" || len(input.Trackers) > 0 || input.Verifier != "" || len(input.Participants) > 0 || len(input.Versions) > 0 || len(input.CustomFieldValues) > 0
}

func resolveProjexProjectID(projectID, spaceID string, format cli.OutputFormat, meta *output.Meta) (string, bool) {
	if projectID != "" && spaceID != "" && projectID != spaceID {
		exitWithError(format, meta, "PARAM_INVALID", "param", false, "project_id and space_id must match when both are set")
		return "", false
	}
	if projectID != "" {
		return projectID, true
	}
	return spaceID, true
}

func resolveAssignedToSelf(ctx context.Context, client *httpx.Client, assignedTo *string, format cli.OutputFormat, meta *output.Meta) bool {
	if *assignedTo != "self" {
		return true
	}
	currentUser, errDetail := orgdomain.GetCurrentUser(ctx, client)
	if errDetail != nil {
		os.Exit(cli.WriteError(errDetail, meta, format))
		return false
	}
	if currentUser.UserID == "" {
		exitWithError(format, meta, "CURRENT_USER_UNRESOLVED", "general", false, "failed to resolve current user id from org current response")
		return false
	}
	*assignedTo = currentUser.UserID
	return true
}

func resolveTextValue(cmd *cobra.Command, textFlag, textValue, fileValue, label string, format cli.OutputFormat, meta *output.Meta) (string, bool) {
	textChanged := cmd.Flags().Changed(textFlag)
	fileChanged := fileValue != ""
	if textChanged && fileChanged {
		exitWithError(format, meta, "PARAM_INVALID", "param", false, label+" and "+label+"_file cannot be used together")
		return "", false
	}
	if !fileChanged {
		return textValue, true
	}
	content, errDetail := readWorkitemTextFile(fileValue)
	if errDetail != nil {
		os.Exit(cli.WriteError(errDetail, meta, format))
		return "", false
	}
	return content, true
}

func validateFormatType(value string, format cli.OutputFormat, meta *output.Meta) bool {
	switch value {
	case "", "RICHTEXT", "MARKDOWN":
		return true
	default:
		exitWithError(format, meta, "PARAM_INVALID", "param", false, "format_type must be one of RICHTEXT or MARKDOWN")
		return false
	}
}

func csvFlag(cmd *cobra.Command, name, value string, format cli.OutputFormat, meta *output.Meta) ([]string, bool) {
	if !cmd.Flags().Changed(name) {
		return nil, true
	}
	values := strings.Split(value, ",")
	out := make([]string, 0, len(values))
	for _, part := range values {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			exitWithError(format, meta, "PARAM_INVALID", "param", false, name+" must be a non-empty comma-separated list")
			return nil, false
		}
		out = append(out, trimmed)
	}
	return out, true
}

func mergeCustomFields(pairs []string, rawJSON string, format cli.OutputFormat, meta *output.Meta) (map[string]any, bool) {
	merged := map[string]any{}
	if rawJSON != "" {
		if err := json.Unmarshal([]byte(rawJSON), &merged); err != nil {
			exitWithError(format, meta, "PARAM_INVALID", "param", false, "custom_fields_json must be a JSON object")
			return nil, false
		}
		if merged == nil {
			exitWithError(format, meta, "PARAM_INVALID", "param", false, "custom_fields_json must be a JSON object")
			return nil, false
		}
	}
	for _, pair := range pairs {
		key, value, ok := strings.Cut(pair, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			exitWithError(format, meta, "PARAM_INVALID", "param", false, "custom_field must use key=value")
			return nil, false
		}
		if _, exists := merged[key]; exists {
			exitWithError(format, meta, "PARAM_INVALID", "param", false, "custom_field duplicates key "+key)
			return nil, false
		}
		merged[key] = value
	}
	if len(merged) == 0 {
		return nil, true
	}
	return merged, true
}

func readWorkitemTextFile(path string) (string, *output.ErrorDetail) {
	base := filepath.Base(path)
	info, err := os.Lstat(path)
	if err != nil {
		return "", workitemFileError(base, "cannot read file")
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return "", workitemFileError(base, "file must be a regular file")
	}
	if info.Size() == 0 {
		return "", workitemFileError(base, "file is empty")
	}
	if info.Size() > maxWorkitemTextFileSize {
		return "", workitemFileError(base, "file exceeds 1MiB")
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return "", workitemFileError(base, "cannot read file")
	}
	if !utf8.Valid(body) {
		return "", workitemFileError(base, "file must be UTF-8")
	}
	return string(body), nil
}

func workitemFileError(base, reason string) *output.ErrorDetail {
	return &output.ErrorDetail{Code: "FILE_READ_FAILED", Category: "general", Retryable: false, Message: "failed to read " + base + ": " + reason}
}
