package docker

import (
	"github.com/croissong/releasechecker/pkg/log"
	"github.com/croissong/releasechecker/pkg/provider"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/mitchellh/mapstructure"
)

type Docker struct {
	config *config
}

type config struct {
	Repo string
}

func (_ Docker) NewProvider(conf map[string]interface{}) (provider.Provider, error) {
	config, err := validateConfig(conf)
	if err != nil {
		return nil, err
	}
	docker := Docker{config: config}
	log.Logger.Debugf("Using %#v", config)
	return &docker, nil
}

func (docker Docker) GetVersion() (string, error) {
	return "", nil
}

func (docker Docker) GetVersions() ([]string, error) {
	tags, err := crane.ListTags(docker.config.Repo)
	if err != nil {
		return nil, err
	}
	var versions []string
	for _, version := range tags {
		versions = append(versions, version)
	}
	log.Logger.Debugf("Versions: %+v", versions)
	return versions, nil
}

func validateConfig(conf map[string]interface{}) (*config, error) {
	var config config
	if err := mapstructure.Decode(conf, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
