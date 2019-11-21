package providers

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/croissong/releasechecker/pkg/config"
	"github.com/croissong/releasechecker/pkg/log"
	"github.com/croissong/releasechecker/pkg/util"
	"github.com/mitchellh/mapstructure"
	"os/exec"
	"strings"
)

type command struct {
	Command string
}

func NewCommand(config map[string]interface{}) (provider, error) {
	var command command
	if err := mapstructure.Decode(config, &command); err != nil {
		return nil, err
	}
	if command.Command == "" {
		return nil, errors.New(fmt.Sprintf("Missing field 'command' in config"))
	}
	log.Logger.Debugf("%#v", command)
	return &command, nil
}

func (cmd command) GetVersion() (string, error) {
	sourceCmd := exec.Command("bash", "-c", fmt.Sprintf("set -o pipefail; %s", cmd.Command))
	log.Logger.Debug("Running source cmd: ", sourceCmd.String())
	var out bytes.Buffer
	sourceCmd.Stdout = &out
	sourceCmd.Stderr = &out
	err := sourceCmd.Run()
	if err != nil {
		errMessage := fmt.Sprintf("Command err: %s - %s", err, out.String())
		if config.Config.InitSources {
			log.Logger.Infof("Ignoring cmd err due to 'initSouces' set. (%s)", errMessage)
			return "", nil
		} else {
			return "", errors.New(errMessage)
		}
	}
	log.Logger.Debugf("Got source version: %s", out.String())
	version := util.StripWhitespace(out.String())
	return version, nil
}

func (cmd command) GetVersions() ([]string, error) {
	sourceCmd := exec.Command("bash", "-c", fmt.Sprintf("set -o pipefail; %s", cmd.Command))
	log.Logger.Debug("Running source cmd: ", sourceCmd.String())
	var out bytes.Buffer
	sourceCmd.Stdout = &out
	sourceCmd.Stderr = &out
	err := sourceCmd.Run()
	if err != nil {
		errMessage := fmt.Sprintf("Command err: %s - %s", err, out.String())
		if config.Config.InitSources {
			log.Logger.Infof("Ignoring cmd err due to 'initSouces' set. (%s)", errMessage)
			return nil, nil
		} else {
			return nil, errors.New(errMessage)
		}
	}
	log.Logger.Debugf("Got source version: %s", out.String())
	versions := strings.Fields(util.StripWhitespace(out.String()))
	return versions, nil
}
