package hooks

import (
	"errors"
	"fmt"
	"github.com/cavaliercoder/grab"
	"github.com/croissong/releasechecker/pkg/log"
	"github.com/croissong/releasechecker/pkg/util"
	"github.com/mholt/archiver"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type downloader struct {
	config *config
	client *grab.Client
}

type config struct {
	Url     string
	Dest    string
	Github  githubConfig
	Extract extractConfig
	Chmod   os.FileMode
}

type extractConfig struct {
	File string
}

type githubConfig struct {
	Repo  string
	Asset string
}

func NewDownloader(conf map[string]interface{}) (hook, error) {
	config, err := validateConfig(conf)
	if err != nil {
		return nil, err
	}
	downloader := downloader{config: config, client: grab.NewClient()}
	return &downloader, nil
}

func (downloader downloader) Run(version string) error {
	url, err := downloader.buildUrl(version)
	if err != nil {
		return err
	}

	tmpDir, err := ioutil.TempDir("", "releasechecker")
	if err != nil {
		log.Logger.Fatal(err)
	}
	defer os.RemoveAll(tmpDir) // clean up

	tmpFilePath, err := downloader.download(url, tmpDir)
	if err != nil {
		log.Logger.Fatal(err)
	}

	extractConf := downloader.config.Extract
	if extractConf != (extractConfig{}) {
		downloader.extract(tmpFilePath, tmpDir)
		tmpFilePath = filepath.Join(tmpDir, downloader.config.Extract.File)
	}

	targetPath := os.ExpandEnv(downloader.config.Dest)
	log.Logger.Debugf("Copy file to %s", targetPath)
	err = util.CopyFile(tmpFilePath, targetPath)
	if err != nil {
		return err
	}
	if downloader.config.Chmod != 0 {
		log.Logger.Debugf("Chmod to %s", downloader.config.Chmod)
		err = os.Chmod(targetPath, downloader.config.Chmod)
		if err != nil {
			return err
		}
	}
	return nil
}

func (downloader downloader) download(url string, destDir string) (string, error) {
	urlParts := strings.Split(url, "/")
	dest := filepath.Join(destDir, urlParts[len(urlParts)-1])
	log.Logger.Infof("Downloading %s to %s ...", url, dest)
	req, _ := grab.NewRequest(dest, url)
	resp := downloader.client.Do(req)
	log.Logger.Infof("Response status: %v", resp.HTTPResponse.Status)

	// start UI loop
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			log.Logger.Infof("Transferred %.2f/%.2fmb (%.0f%%)",
				float64(resp.BytesComplete())/1000000,
				float64(resp.Size)/1000000,
				100*resp.Progress())

		case <-resp.Done:
			break Loop
		}
	}

	if err := resp.Err(); err != nil {
		log.Logger.Errorf("Download failed: %v", err)
		return "", err
	}

	log.Logger.Infof("Download saved to %s", resp.Filename)
	return resp.Filename, nil
}

func (downloader downloader) extract(archive string, dest string) error {
	log.Logger.Infof("Extracting archive %s to %s", archive, dest)
	err := archiver.Unarchive(archive, dest)
	if err != nil {
		log.Logger.Fatal(err)
	}
	return nil
}

const githubUrlTmpl = "https://github.com/{{.Repo}}/releases/download/{{.Version}}/{{.Asset}}"

func (downloader downloader) buildUrl(version string) (string, error) {
	if downloader.config.Url != "" {
		templateData := struct {
			Version string
		}{
			Version: version,
		}
		url, err := util.RenderTemplate(downloader.config.Url, templateData)
		if err != nil {
			return "", err
		}
		return url, nil
	}

	githubConf := downloader.config.Github
	templateData := struct {
		Version string
		Repo    string
		Asset   string
	}{
		Version: version,
		Repo:    githubConf.Repo,
		Asset:   githubConf.Asset,
	}
	url, err := util.RenderTemplate(githubUrlTmpl, templateData)
	if err != nil {
		return "", err
	}
	return url, nil
}

func validateConfig(conf map[string]interface{}) (*config, error) {
	var config config
	if err := mapstructure.Decode(conf, &config); err != nil {
		return nil, err
	}
	log.Logger.Debugf("%#v", config)
	if config.Dest == "" {
		return nil, errors.New(fmt.Sprintf("Missing field 'dest' in config"))
	}

	if config.Url == "" && (config.Github.Repo == "" && config.Github.Asset == "") {
		return nil, errors.New(fmt.Sprintf("Invalid config: Missing field 'url' or 'github'"))
	}
	return &config, nil
}
