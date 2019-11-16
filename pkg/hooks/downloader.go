package hooks

import (
	"bytes"
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
	"text/template"
	"time"
)

type downloader struct {
	config config
	client *grab.Client
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
	downloader := downloader{config: config, client: grab.NewClient()}
	log.Logger.Debugf("%#v", downloader)
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
	err = util.CopyFile(tmpFilePath, targetPath)
	if err != nil {
		return err
	}
	return nil
}

func (downloader downloader) download(url string, dest string) (string, error) {
	req, _ := grab.NewRequest(dest, url)
	log.Logger.Info("Downloading %v...", req.URL())
	resp := downloader.client.Do(req)
	log.Logger.Info("Response status: %v", resp.HTTPResponse.Status)

	// start UI loop
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			log.Logger.Info("%.02f%% complete", resp.Progress())

		case <-resp.Done:
			break Loop
		}
	}

	if err := resp.Err(); err != nil {
		log.Logger.Errorf("Download failed: %v", err)
		return "", err
	}

	filePath := filepath.Join(dest, resp.Filename)
	log.Logger.Infof("Download saved to %s/%v")
	return filePath, nil
}

func (downloader downloader) extract(archive string, dest string) error {
	log.Logger.Infof("Extracting archive %s to %s", archive, dest)
	err := archiver.Unarchive(archive, dest)
	if err != nil {
		log.Logger.Fatal(err)
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
