package flagmeta

import "github.com/spf13/cobra"

const RequiredFlagAnnotation = "yunxiao_required_flag"

func MustMarkRequired(cmd *cobra.Command, names ...string) {
	for _, name := range names {
		if err := cmd.Flags().SetAnnotation(name, RequiredFlagAnnotation, []string{"true"}); err != nil {
			panic(err)
		}
	}
}
