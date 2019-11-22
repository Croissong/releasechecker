package providers

import (
	"github.com/croissong/releasechecker/pkg/log"
	"github.com/croissong/releasechecker/pkg/util"
	"github.com/mitchellh/mapstructure"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
)

type yaml struct {
	config *yamlConfig
}

type yamlConfig struct {
	Path string
	Url  string
}

func NewYaml(conf map[string]interface{}) (provider, error) {
	config, err := validateConfig(conf)
	if err != nil {
		return nil, err
	}
	yaml := yaml{config: config}
	return &yaml, nil
}

func (yaml yaml) GetVersion() (string, error) {
	resp, err := http.Get(yaml.config.Url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(body)
	version, err := execYq(yaml.config.Path, bodyString)
	if err != nil {
		return "", err
	}
	return version, nil
}

func (yaml yaml) GetVersions() ([]string, error) {
	return nil, nil
}

func execYq(path string, input string) (string, error) {
	cmd := exec.Command("yq", "-r", path)
	log.Logger.Debug("Running yq cmd: ", cmd.String())
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Logger.Error(err)
		return "", err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, input)
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Logger.Error(err)
		return "", err
	}
	return util.StripWhitespace(string(out)), nil

}

func validateConfig(conf map[string]interface{}) (*yamlConfig, error) {
	var config yamlConfig
	if err := mapstructure.Decode(conf, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
