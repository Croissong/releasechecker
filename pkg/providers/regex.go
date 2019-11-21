package providers

import (
	"github.com/croissong/releasechecker/pkg/log"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"net/http"
	"regexp"
)

type regex struct {
	Regex string
	Url   string
}

func NewRegex(config map[string]interface{}) (provider, error) {
	var regex regex
	if err := mapstructure.Decode(config, &regex); err != nil {
		return nil, err
	}
	log.Logger.Debugf("%#v", regex)
	return &regex, nil
}

func (regex regex) GetVersion() (string, error) {
	return "", nil
}

func (regex regex) GetVersions() ([]string, error) {
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
	matches := versionRegex.FindAllStringSubmatch(bodyString, -1)
	var versions []string
	for _, match := range matches {
		versions = append(versions, match[1])
	}
	log.Logger.Debug("%s", versions)
	return versions, nil
}
