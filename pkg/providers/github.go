package providers

import (
	"encoding/json"
	"fmt"
	"github.com/croissong/releasechecker/pkg/log"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"net/http"
)

const urlTemplate = "https://api.github.com/repos/%s/releases"

type github struct {
	Repo string
}

type releaseDto struct {
	TagName string `json:"tag_name"`
}

func NewGithub(config map[string]interface{}) (provider, error) {
	var github github
	if err := mapstructure.Decode(config, &github); err != nil {
		return nil, err
	}
	log.Logger.Debugf("%#v", github)
	return &github, nil
}

func (github github) GetVersion() (string, error) {
	return "", nil
}

func (github github) GetVersions() ([]string, error) {
	url := fmt.Sprintf(urlTemplate, github.Repo)
	log.Logger.Debugf("Fetching github releases from %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var releases []releaseDto
	if err = json.Unmarshal(body, &releases); err != nil {
		return nil, err
	}
	log.Logger.Debugf("Fetched releases: %#v", releases)

	var versions []string
	for _, release := range releases {
		versions = append(versions, release.TagName)
	}
	log.Logger.Debugf("Versions: %#v", versions)
	return versions, nil
}
