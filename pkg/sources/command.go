package sources

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/croissong/releasechecker/pkg/log"
	"github.com/croissong/releasechecker/pkg/util"
	"github.com/mitchellh/mapstructure"
	"os/exec"
	"strings"
)

type command struct {
	Command string
}

func NewCommand(config map[string]interface{}) (source, error) {
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
		if strings.Contains(out.String(), "command not found") {
			return "", nil
		}
		return "", err
	}
	log.Logger.Debugf("Got source version: %s", out.String())
	version := util.StripWhitespace(out.String())
	return version, nil
}
