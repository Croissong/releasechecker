package providers

import (
	"github.com/croissong/releasechecker/pkg/log"
	"io/ioutil"
	"net/http"
	"regexp"
)

func GetVersions() []string {
	regex := `.*>terraform_(.*)<.*`
	url := "https://releases.hashicorp.com/terraform/"
	resp, err := http.Get(url)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	bodyString := string(body)
	var versionRegex = regexp.MustCompile(regex)
	matches := versionRegex.FindAllStringSubmatch(bodyString, -1)
	var versions []string
	for _, match := range matches {
		versions = append(versions, match[1])
	}
	log.Logger.Debug("%s", versions)
	return versions
}
