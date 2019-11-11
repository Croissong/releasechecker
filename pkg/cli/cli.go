package cli

import (
	"github.com/croissong/releasechecker/pkg/config"
	"github.com/croissong/releasechecker/pkg/hooks"
	"github.com/croissong/releasechecker/pkg/log"
	"github.com/croissong/releasechecker/pkg/providers"
	"github.com/croissong/releasechecker/pkg/sources"
	"github.com/croissong/releasechecker/pkg/versions"
	"github.com/spf13/cobra"
	"os"
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

		provider := providers.GetProvider(entries[0].Provider)
		versionStrings := provider.GetVersions()
		latestVersion, err := versions.GetLatestVersion(versionStrings)
		if err != nil {
			log.Logger.Fatal(err)
		}
		log.Logger.Info("Latest version ", latestVersion)

		source := sources.GetSource(entries[0].Source)
		currentVersion, err := source.GetVersion()
		if err != nil {
			log.Logger.Fatal(err)
		}

		if currentVersion == nil {
			log.Logger.Infof("No current version for %s detected", name)
			hooks.RunHooks(latestVersion.Original(), entries[0].Hooks)
			return
		}

		log.Logger.Infof("The current version for %s is %s", name, currentVersion)

		if versions.IsNewer(latestVersion, currentVersion) {
			log.Logger.Info("Newer version detected")
			hooks.RunHooks(latestVersion.Original(), entries[0].Hooks)
		} else {
			log.Logger.Info("No new version detected")
		}
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
