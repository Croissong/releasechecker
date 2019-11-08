package cli

import (
	"fmt"
	"github.com/croissong/releasechecker/pkg/config"
	"github.com/croissong/releasechecker/pkg/log"
	"github.com/croissong/releasechecker/pkg/providers"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "releasechecker",
	Short: "Check relase",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		entries := config.Config.Entries
		name := entries[0].Name
		source := entries[0].Source
		sourceCmd := exec.Command("bash", "-c", fmt.Sprintf("%s", source))
		log.Logger.Debug("Running source cmd: ", sourceCmd.String())
		out, err := sourceCmd.Output()
		if err != nil {
			log.Logger.Error(err)
		}
		log.Logger.Infof("The version for %s is %s", name, out)
		providers.GetVersion()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Logger.Error(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(log.InitLogger, config.InitConfig)
	rootCmd.PersistentFlags().StringVar(&config.CfgFile, "config", "", "config file (default is $HOME/.releasechecker.yaml)")
}
