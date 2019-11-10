package providers

import (
	"github.com/croissong/releasechecker/pkg/log"
	"io/ioutil"
	"net/http"
	"regexp"
)

type regex struct {
	versionRegex string
}

func (cmd regex) Init(providerConfig map[string]interface{}) provider {
	cmd.versionRegex = providerConfig["regex"].(string)
	return cmd
}

func (regex regex) GetVersions() []string {
	var versionRegex = regexp.MustCompile(regex.versionRegex)
	url := "https://releases.hashicorp.com/terraform/"
	resp, err := http.Get(url)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	bodyString := string(body)
	matches := versionRegex.FindAllStringSubmatch(bodyString, -1)
	var versions []string
	for _, match := range matches {
		versions = append(versions, match[1])
	}
	log.Logger.Debug("%s", versions)
	return versions
}
