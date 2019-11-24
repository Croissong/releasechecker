package cli

import (
	"errors"
	"github.com/croissong/releasechecker/pkg/config"
	"github.com/croissong/releasechecker/pkg/hooks"
	"github.com/croissong/releasechecker/pkg/log"
	"github.com/croissong/releasechecker/pkg/provider"
	"github.com/croissong/releasechecker/pkg/provider/command"
	"github.com/croissong/releasechecker/pkg/provider/docker"
	"github.com/croissong/releasechecker/pkg/provider/github"
	"github.com/croissong/releasechecker/pkg/provider/regex"
	"github.com/croissong/releasechecker/pkg/provider/yaml"
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
		log.Logger.Error(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(log.ConfigureLogger, config.InitConfig)
	rootCmd.PersistentFlags().StringVar(&config.CfgFile, "config", "", "config file (default is $HOME/.releasechecker.yaml)")
}

var providers = map[string]provider.Provider{
	"command": command.Command{},
	"github":  github.Github{},
	"regex":   regex.Regex{},
	"docker":  docker.Docker{},
	"yaml":    yaml.Yaml{},
}

func checkReleases() {
	entries := config.Config.Entries
	for _, entry := range entries {
		name := entry.Name
		log.Logger.Info("Checking version for ", name)

		upstreamVersion, err := getUpstreamVersion(entry.Upstream)
		if err != nil {
			log.Logger.Fatal(err)
		}

		downstreamVersion, err := getDownstreamVersion(entry.Downstream)
		if err != nil {
			log.Logger.Fatal(err)
		}

		if downstreamVersion == nil {
			log.Logger.Infof("No downstream version for %s detected", name)
			if err = hooks.RunHooks(upstreamVersion.Original(), downstreamVersion.Original(), entry.Hooks); err != nil {
				log.Logger.Fatal(err)
			}
			return
		}

		log.Logger.Infof("The current version for %s is %s", name, downstreamVersion)

		if provider.IsNewerVersion(upstreamVersion, downstreamVersion) {
			log.Logger.Info("Newer version detected")
			if err = hooks.RunHooks(upstreamVersion.Original(), downstreamVersion.Original(), entry.Hooks); err != nil {
				log.Logger.Fatal(err)
			}
		} else {
			log.Logger.Info("No new version detected")
		}
	}
}

func getUpstreamVersion(upstreamConfig map[string]interface{}) (*ver.Version, error) {
	if len(upstreamConfig) == 0 {
		return nil, errors.New("Missing upstream configuration")
	}
	upstream, err := provider.GetProvider(providers, upstreamConfig)
	if err != nil {
		log.Logger.Error(err)
		return nil, errors.New("Error getting upstream provider")
	}
	latestVersion, err := provider.GetLatestVersion(upstream)
	if err != nil {
		log.Logger.Error(err)
		return nil, errors.New("Invalid upstream version")
	}
	log.Logger.Info("Upstream version is ", latestVersion)
	return latestVersion, nil
}

func getDownstreamVersion(downstreamConfig map[string]interface{}) (*ver.Version, error) {
	if len(downstreamConfig) == 0 {
		return nil, errors.New("Missing downstream configuration")
	}
	downstream, err := provider.GetProvider(providers, downstreamConfig)
	if err != nil {
		log.Logger.Error(err)
		return nil, errors.New("Error getting downstream provider")
	}
	version, err := downstream.GetVersion()
	if err != nil {
		log.Logger.Error(err)
		return nil, errors.New("Error getting downstream version")
	}
	return version, nil
}
