package cmd

import (
	"os"

	"github.com/fmterrorf/drdrei/internal"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "drdrei paths...",
	Short: "Detect outdated Terraform module sources",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			if err := cmd.Help(); err != nil {
				return err
			}
			os.Exit(0)
		}
		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}
		ignorePaths, err := cmd.Flags().GetStringArray("ignorePaths")
		if err != nil {
			return err
		}
		printAsJSON, err := cmd.Flags().GetBool("json")
		if err != nil {
			return err
		}
		internal.RunAudit(args, recursive, ignorePaths, printAsJSON)
		return nil
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.drdrei.yaml)")
	rootCmd.Flags().BoolP("recursive", "r", false, "Run audit recursively to given paths")
	rootCmd.Flags().StringArrayP("ignorePaths", "i", []string{".terraform", ".git"}, "Path to ignore")
	rootCmd.Flags().Bool("json", false, "Show result as JSON")
}
