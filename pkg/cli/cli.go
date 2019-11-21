package cli

import (
	"errors"
	"github.com/croissong/releasechecker/pkg/config"
	"github.com/croissong/releasechecker/pkg/hooks"
	. "github.com/croissong/releasechecker/pkg/log"
	"github.com/croissong/releasechecker/pkg/providers"
	"github.com/croissong/releasechecker/pkg/versions"
	ver "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "releasechecker",
	Short: "Check upstream releases, compare them with downstream and execute hooks on changes.",
	Long:  `Check upstream releases, compare them with downstream and execute hooks on changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		checkReleases()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		Logger.Error(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(InitLogger, config.InitConfig)
	rootCmd.PersistentFlags().StringVar(&config.CfgFile, "config", "", "config file (default is $HOME/.releasechecker.yaml)")
}

func checkReleases() {
	entries := config.Config.Entries
	for _, entry := range entries {
		name := entry.Name
		Logger.Info("Checking version for ", name)

		upstreamVersion, err := getUpstreamVersion(entry.Upstream)
		if err != nil {
			Logger.Fatal(err)
		}

		downstreamVersion, err := getDownstreamVersion(entry.Downstream)
		if err != nil {
			Logger.Fatal(err)
		}

		if downstreamVersion == nil {
			Logger.Infof("No downstream version for %s detected", name)
			if err = hooks.RunHooks(upstreamVersion.Original(), entry.Hooks); err != nil {
				Logger.Fatal(err)
			}
			return
		}

		Logger.Infof("The current version for %s is %s", name, downstreamVersion)

		if versions.IsNewer(upstreamVersion, downstreamVersion) {
			Logger.Info("Newer version detected")
			if err = hooks.RunHooks(upstreamVersion.Original(), entry.Hooks); err != nil {
				Logger.Fatal(err)
			}
		} else {
			Logger.Info("No new version detected")
		}
	}
}

func getUpstreamVersion(upstreamConfig map[string]interface{}) (*ver.Version, error) {
	if len(upstreamConfig) == 0 {
		return nil, errors.New("Missing upstream configuration")
	}
	upstream, err := providers.GetProvider(upstreamConfig)
	if err != nil {
		Logger.Error(err)
		return nil, errors.New("Error getting upstream provider")
	}
	versionStrings, err := upstream.GetVersions()
	if err != nil {
		Logger.Error(err)
		return nil, errors.New("Error getting upsteam versions")
	}
	latestVersion, err := versions.GetLatestVersion(versionStrings)
	if err != nil {
		Logger.Error(err)
		return nil, errors.New("Invalid upstream version")
	}
	Logger.Info("Upstream version is ", latestVersion)
	return latestVersion, nil
}

func getDownstreamVersion(downstreamConfig map[string]interface{}) (*ver.Version, error) {
	if len(downstreamConfig) == 0 {
		return nil, errors.New("Missing downstream configuration")
	}
	downstream, err := providers.GetProvider(downstreamConfig)
	if err != nil {
		Logger.Error(err)
		return nil, errors.New("Error getting downstream provider")
	}
	currentVersionString, err := downstream.GetVersion()
	if err != nil {
		Logger.Error(err)
		return nil, errors.New("Error getting downstream version")
	}

	if currentVersionString == "" {
		return nil, nil
	}

	currentVersion, err := ver.NewVersion(currentVersionString)
	if err != nil {
		Logger.Error(err)
		return nil, errors.New("Invalid downstream version")
	}
	return currentVersion, nil
}
