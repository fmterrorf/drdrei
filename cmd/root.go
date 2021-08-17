package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/fmterrorf/drdrei/internal"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "drdrei paths...",
	Short: "Detect outdated Terraform module sources",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := cmd.Help(); err != nil {
				log.Fatal(err)
			}
			os.Exit(0)
		}
		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			log.Fatal(err)
		}
		ignorePaths, err := cmd.Flags().GetStringArray("ignorePaths")
		if err != nil {
			log.Fatal(err)
		}
		internal.RunAudit(args, recursive, ignorePaths)
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.drdrei.yaml)")
	rootCmd.Flags().BoolP("recursive", "r", false, "Run audit recursively to given paths")
	rootCmd.Flags().StringArrayP("ignorePaths", "i", []string{".terraform", ".git"}, "Path to ignore")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.SetConfigName(".drdrei")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
