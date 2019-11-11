package hooks

import (
	"bytes"
	"fmt"
	"github.com/croissong/releasechecker/pkg/log"
	"github.com/croissong/releasechecker/pkg/util"
	"github.com/mholt/archiver"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type downloaderConfig struct {
	Url     string
	Dest    string
	Extract extractConfig
}

type extractConfig struct {
	File string
}

type downloader struct {
	urlTemplate string
	targetPath  string
	extract     bool
	extractFile string
}

func (downloader downloader) Init(hookConfig map[string]interface{}) (hook, error) {
	var config downloaderConfig
	if err := mapstructure.Decode(hookConfig, &config); err != nil {
		return nil, err
	}
	downloader.urlTemplate = config.Url
	downloader.targetPath = os.ExpandEnv(config.Dest)
	if config.Extract != (extractConfig{}) {
		downloader.extract = true
		downloader.extractFile = config.Extract.File
	} else {
		downloader.extract = false
	}
	return downloader, nil
}

func (downloader downloader) Run(version string) error {
	url, err := downloader.buildUrl(version)
	if err != nil {
		return err
	}
	urlParts := strings.Split(url, "/")
	downloadPath := urlParts[len(urlParts)-1]

	tmpDir, err := ioutil.TempDir("", fmt.Sprintf("%s*", downloadPath))
	defer os.RemoveAll(tmpDir) // clean up

	log.Logger.Info("Downloading ", url)
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	tmpfn := filepath.Join(tmpDir, downloadPath)
	log.Logger.Debug("Downloading to ", tmpfn)
	if err != nil {
		log.Logger.Fatal(err)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Logger.Fatal(err)
		return err
	}

	if err := ioutil.WriteFile(tmpfn, body, 0666); err != nil {
		log.Logger.Fatal(err)
		return err
	}

	if downloader.extract {
		log.Logger.Info("Extracting archive ", tmpfn)
		err = archiver.Unarchive(tmpfn, tmpDir)
		if err != nil {
			log.Logger.Fatal(err)
		}
		extractedFilePath := filepath.Join(tmpDir, downloader.extractFile)
		log.Logger.Debugf("Renaming %s to %s", extractedFilePath, downloader.targetPath)
		err := util.CopyFile(extractedFilePath, downloader.targetPath)
		if err != nil {
			return err
		}
	} else {
		err := util.CopyFile(tmpfn, downloader.targetPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (downloader downloader) buildUrl(version string) (string, error) {
	tmpl, err := template.New("urlTemplate").Parse(downloader.urlTemplate)
	if err != nil {
		return "", err
	}
	data := struct {
		Version string
	}{
		Version: version,
	}
	var tpl bytes.Buffer
	tmpl.Execute(&tpl, data)

	return tpl.String(), nil
}
