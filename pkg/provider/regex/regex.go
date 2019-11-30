package regex

import (
	"github.com/croissong/releasechecker/pkg/log"
	"github.com/croissong/releasechecker/pkg/provider"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"net/http"
	"regexp"
)

type Regex struct {
	Regex string
	Url   string
}

func (_ Regex) NewProvider(config map[string]interface{}) (provider.Provider, error) {
	var regex Regex
	if err := mapstructure.Decode(config, &regex); err != nil {
		return nil, err
	}
	log.Logger.Debugf("%#v", regex)
	return &regex, nil
}

func (regex Regex) GetVersion() (string, error) {
	versions, err := regex.GetVersions()
	if err != nil {
		return "", err
	}
	return versions[0], nil
}

func (regex Regex) GetVersions() ([]string, error) {
	var versionRegex = regexp.MustCompile(regex.Regex)
	resp, err := http.Get(regex.Url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	bodyString := string(body)
	log.Logger.Debug(bodyString)
	matches := versionRegex.FindAllStringSubmatch(bodyString, -1)
	var versions []string
	for _, match := range matches {
		version := match[1]
		versions = append(versions, version)
	}
	log.Logger.Debugf("%s", versions)
	return versions, nil
}
