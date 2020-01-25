package hooks

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/croissong/releasechecker/pkg/log"
	"github.com/croissong/releasechecker/pkg/util"
	cmdutil "github.com/croissong/releasechecker/pkg/util/cmd"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type gitHook struct {
	config  *gitConfig
	repo    *git.Repository
	repoDir string
}

type gitConfig struct {
	Repo   string
	Branch string
	Change changeConfig
	Commit commitConfig
}

type commitConfig struct {
	MsgTemplate string
	Branch      string
	Push        bool
	AuthorName  string
	AuthorEmail string
	Tag         string
}

type changeConfig struct {
	Command string
}

func (_ gitHook) NewHook(conf map[string]interface{}) (hook, error) {
	config, err := validateGitConfig(conf)
	if err != nil {
		return nil, err
	}
	repoDir, err := getRepoCacheDir(config)
	if err != nil {
		return nil, err
	}
	gitHook := gitHook{config: config, repoDir: repoDir}
	return &gitHook, nil
}

func (gitHook gitHook) Run(newVersion string, oldVersion string) error {
	conf := gitHook.config
	var repo *git.Repository
	repo, err := gitHook.clone()
	if err == git.ErrRepositoryAlreadyExists {
		log.Logger.Info("Repo already exists")
		repo, err = gitHook.checkout()
	}
	if err != nil {
		return err
	}
	gitHook.repo = repo

	err = gitHook.change(newVersion, oldVersion)
	if err != nil {
		return err
	}
	commit, err := gitHook.commit(newVersion, oldVersion)
	if err != nil {
		return err
	}

	if commit != nil && conf.Commit.Tag != "" {
		err = gitHook.tag(*commit, newVersion, conf.Commit)
		if err != nil {
			return err
		}
	}

	if commit != nil && conf.Commit.Push {
		err = gitHook.push()
		if err != nil {
			log.Logger.Error(err)
			return err
		}
	}
	return nil
}

func (gitHook gitHook) clone() (*git.Repository, error) {
	url := gitHook.config.Repo
	var repo *git.Repository
	log.Logger.Infof("git clone %s %s", url, gitHook.repoDir)
	repo, err := git.PlainClone(gitHook.repoDir, false, &git.CloneOptions{
		URL:   url,
		Depth: 2,
	})
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (gitHook gitHook) checkout() (*git.Repository, error) {
	repo, err := git.PlainOpen(gitHook.repoDir)
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	worktree, err := repo.Worktree()
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	log.Logger.Debug("Checking out master")
	err = worktree.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash("master"),
	})
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}

	log.Logger.Debug("Fetching origin")
	err = repo.Fetch(&git.FetchOptions{
		RemoteName: git.DefaultRemoteName,
		Depth:      2,
		Force:      true,
	})
	if err == git.NoErrAlreadyUpToDate {
		log.Logger.Debug("Already up-to-date")
	} else if err != nil {
		log.Logger.Error(err)
		return nil, err
	}

	log.Logger.Debug("Resetting to origin")
	var remoteHash plumbing.Hash
	remoteRef, err := repo.Reference(plumbing.ReferenceName("refs/remotes/origin/master"), true)
	if err != nil {
		return nil, err
	}

	remoteHash = remoteRef.Hash()
	err = worktree.Reset(&git.ResetOptions{Commit: remoteHash, Mode: git.HardReset})
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	return repo, nil
}

func (gitHook gitHook) change(newVersion string, oldVersion string) error {
	commandTemplate := gitHook.config.Change.Command
	templateData := struct {
		NewVersion string
		OldVersion string
	}{
		NewVersion: newVersion,
		OldVersion: oldVersion,
	}
	command, err := util.RenderTemplate(commandTemplate, templateData)
	log.Logger.Debug(command)
	_, err = cmdutil.RunCmd(command, cmdutil.CmdOptions{Dir: gitHook.repoDir})
	if err != nil {
		log.Logger.Error(err)
		return err
	}
	return nil
}

func (gitHook gitHook) commit(newVersion string, oldVersion string) (*plumbing.Hash, error) {
	commitConf := gitHook.config.Commit
	worktree, err := gitHook.repo.Worktree()
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	worktree.Add(".")
	status, _ := worktree.Status()
	if status.IsClean() {
		log.Logger.Warn("Nothing to commit")
		return nil, nil
	}

	templateData := struct {
		NewVersion string
		OldVersion string
	}{
		NewVersion: newVersion,
		OldVersion: oldVersion,
	}
	message, err := util.RenderTemplate(commitConf.MsgTemplate, templateData)
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	commit, err := worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  commitConf.AuthorName,
			Email: commitConf.AuthorEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	obj, err := gitHook.repo.CommitObject(commit)
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}
	log.Logger.Debugf("Committed: %s", obj)
	return &commit, nil
}

func (gitHook gitHook) tag(commit plumbing.Hash, newVersion string, commitConf commitConfig) error {
	templateData := struct {
		NewVersion string
	}{
		NewVersion: newVersion,
	}
	tag, err := util.RenderTemplate(commitConf.Tag, templateData)
	if err != nil {
		log.Logger.Error(err)
		return err
	}
	_, err = gitHook.repo.CreateTag(tag, commit, nil)
	if err != nil {
		log.Logger.Error(err)
		return err
	}
	log.Logger.Debugf("Created tag: %s on commit %s", tag, commit)
	return err
}

func (gitHook gitHook) push() error {
	log.Logger.Debug("Pushing")
	err := gitHook.repo.Push(&git.PushOptions{
		RefSpecs: []config.RefSpec{
			config.RefSpec("refs/heads/master:refs/heads/master"),
			config.RefSpec("refs/tags/*:refs/tags/*"),
		},
	})
	return err
}

func validateGitConfig(conf map[string]interface{}) (*gitConfig, error) {
	config := gitConfig{
		Branch: "master",
	}
	if err := mapstructure.Decode(conf, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

var protocolRegex = regexp.MustCompile("git@?([^:]*):(.*).git")

func getRepoCacheDir(config *gitConfig) (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	repoDirName := strings.ToLower(protocolRegex.ReplaceAllString(config.Repo, "$1/$2"))
	repoDir := filepath.Join(cacheDir, "releasewatcher/repos", repoDirName)
	return repoDir, nil
}
