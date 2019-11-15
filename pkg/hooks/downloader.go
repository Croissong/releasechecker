package hooks

import (
	"bytes"
	"errors"
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

type downloader struct {
	config config
}

type config struct {
	Url     string
	Dest    string
	Extract extractConfig
}

type extractConfig struct {
	File string
}

func NewDownloader(conf map[string]interface{}) (hook, error) {
	var config config
	if err := mapstructure.Decode(conf, &config); err != nil {
		return nil, err
	}
	if config.Url == "" {
		return nil, errors.New(fmt.Sprintf("Missing field 'url' in config"))
	}
	if config.Dest == "" {
		return nil, errors.New(fmt.Sprintf("Missing field 'dest' in config"))
	}
	downloader := downloader{config: config}
	log.Logger.Debugf("%#v", downloader)
	return &downloader, nil
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
	targetPath := os.ExpandEnv(downloader.config.Dest)

	if downloader.config.Extract.File != "" {
		log.Logger.Info("Extracting archive ", tmpfn)
		err = archiver.Unarchive(tmpfn, tmpDir)
		if err != nil {
			log.Logger.Fatal(err)
		}
		extractedFilePath := filepath.Join(tmpDir, downloader.config.Extract.File)
		log.Logger.Debugf("Renaming %s to %s", extractedFilePath, targetPath)
		err := util.CopyFile(extractedFilePath, targetPath)
		if err != nil {
			return err
		}
	} else {
		err := util.CopyFile(tmpfn, targetPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (downloader downloader) buildUrl(version string) (string, error) {
	tmpl, err := template.New("urlTemplate").Parse(downloader.config.Url)
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
