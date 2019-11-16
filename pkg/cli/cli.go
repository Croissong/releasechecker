package cli

import (
	"github.com/croissong/releasechecker/pkg/config"
	"github.com/croissong/releasechecker/pkg/hooks"
	"github.com/croissong/releasechecker/pkg/log"
	"github.com/croissong/releasechecker/pkg/providers"
	"github.com/croissong/releasechecker/pkg/sources"
	"github.com/croissong/releasechecker/pkg/versions"
	ver "github.com/hashicorp/go-version"
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
		checkReleases()
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

func checkReleases() {
	entries := config.Config.Entries
	for _, entry := range entries {
		name := entry.Name
		log.Logger.Info("Checking version for ", name)
		provider, err := providers.GetProvider(entry.Provider)
		if err != nil {
			log.Logger.Fatal(err)
		}
		versionStrings, err := provider.GetVersions()
		if err != nil {
			log.Logger.Fatal(err)
		}
		latestVersion, err := versions.GetLatestVersion(versionStrings)
		if err != nil {
			log.Logger.Fatal(err)
		}
		log.Logger.Info("Latest version is ", latestVersion)

		source, err := sources.GetSource(entry.Source)
		if err != nil {
			log.Logger.Fatal(err)
		}
		currentVersionString, err := source.GetVersion()
		if err != nil {
			log.Logger.Fatal(err)
		}

		if currentVersionString == "" {
			log.Logger.Infof("No current version for %s detected", name)
			if err = hooks.RunHooks(latestVersion.Original(), entry.Hooks); err != nil {
				log.Logger.Fatal(err)
			}
			return
		}

		currentVersion, err := ver.NewVersion(currentVersionString)
		if err != nil {
			log.Logger.Fatal(err)
		}

		log.Logger.Infof("The current version for %s is %s", name, currentVersion)

		if versions.IsNewer(latestVersion, currentVersion) {
			log.Logger.Info("Newer version detected")
			if err = hooks.RunHooks(latestVersion.Original(), entry.Hooks); err != nil {
				log.Logger.Fatal(err)
			}
		} else {
			log.Logger.Info("No new version detected")
		}
	}
}
