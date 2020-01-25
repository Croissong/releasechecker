package provider

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/croissong/releasechecker/pkg/log"
	"sort"
)

type Provider interface {
	NewProvider(map[string]interface{}) (Provider, error)
	GetVersion() (string, error)
	GetVersions() ([]string, error)
}

func GetProvider(providers map[string]Provider, providerConfig map[string]interface{}) (Provider, error) {
	if providerType, ok := providerConfig["type"]; ok {
		providerType := providerType.(string)
		if provider, ok := providers[providerType]; ok {
			return provider.NewProvider(providerConfig)
		}
		return nil, errors.New(fmt.Sprintf("Provider '%s' not found", providerType))
	}
	return nil, errors.New("Missing 'type' key in provider config")
}

func GetLatestVersion(provider Provider) (*semver.Version, error) {
	vStrings, err := provider.GetVersions()
	if err != nil {
		return nil, err
	}
	var versions []*semver.Version
	for _, vString := range vStrings {
		version, err := semver.NewVersion(vString)
		if err != nil {
			log.Logger.Debugf("Ignoring version %s (%s)", vString, err)
			continue
		}
		versions = append(versions, version)
	}
	sort.Sort(semver.Collection(versions))
	log.Logger.Debug(versions)
	if len(versions) == 0 {
		return nil, errors.New("Found 0 versions")
	}
	latestVersion := versions[len(versions)-1]
	return latestVersion, nil
}
