package main

import (
	"fmt"

	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"changelog-version/repository"
	"changelog-version/utils"

	changelog "github.com/anton-yurchenko/go-changelog"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/afero"
	"golang.org/x/mod/semver"
)

const dateFormat string = `2006-01-02`

type app struct {
	version          string
	updateMajorMinor bool
	changelog        changelogFile
	git              *repository.Repository
}

type changelogFile struct {
	filename string
	parser   *changelog.Parser
	content  *changelog.Changelog
}

func new() (*app, error) {
	v := os.Getenv("VERSION")
	if !semver.IsValid(v) {
		return nil, fmt.Errorf("invalid semantic version (make sure to add a 'v' prefix: vX.X.X)")
	}

	var u bool
	if os.Getenv("UPDATE_TAGS") != "" {
		var err error
		u, err = strconv.ParseBool(os.Getenv("UPDATE_TAGS"))
		if err != nil {
			return nil, utils.Wrap("error parsing UPDATE_TAGS environmental variable: %s", err)
		}
	}

	g, err := repository.New(os.Getenv("GITHUB_TOKEN"), os.Getenv("GITHUB_ACTOR"))
	if err != nil {
		return nil, utils.Wrap("git configuration error: %s", err)
	}

	f := "CHANGELOG.md"
	if x := os.Getenv("CHANGELOG_FILE"); x != "" {
		f = x
	}

	p, err := changelog.NewParser(f)
	if err != nil {
		return nil, utils.Wrap("error initializing changelog parser: %s", err)
	}

	c, err := p.Parse()
	if err != nil {
		return nil, utils.Wrap("error parsing changelog: %s", err)
	}

	return &app{
		version:          v,
		updateMajorMinor: u,
		changelog: changelogFile{
			filename: f,
			parser:   p,
			content:  c,
		},
		git: g,
	}, nil
}

func (a *app) updateChangelog() error {
	url := fmt.Sprintf("https://github.com/%s", os.Getenv("GITHUB_REPOSITORY"))

	releaseURL := fmt.Sprintf("%s/releases/tag/%s", url, a.version)
	if len(a.changelog.content.Releases) > 0 {
		t := a.changelog.content.Releases
		sort.Sort(t)
		if t[len(t)-1].Version != nil {
			releaseURL = fmt.Sprintf("%s/compare/v%s...%s", url, *t[len(t)-1].Version, a.version)
		}
	}

	if _, err := a.changelog.content.CreateReleaseFromUnreleasedWithURL(
		strings.TrimPrefix(a.version, "v"),
		time.Now().Format(dateFormat),
		releaseURL,
	); err != nil {
		return utils.Wrap("error creating release from an unreleased: %s", err)
	}

	if err := a.changelog.content.SetUnreleasedURL(fmt.Sprintf("%s/compare/%s...HEAD", url, a.version)); err != nil {
		return utils.Wrap("error updating unreleased url: %s", err)
	}

	if err := a.changelog.content.SaveToFile(afero.NewOsFs(), a.changelog.filename); err != nil {
		return utils.Wrap("error saving changelog to file: %s", err)
	}

	return nil
}

func (a *app) commit() (plumbing.Hash, error) {
	return a.git.Commit(strings.TrimPrefix(a.version, "v"))
}

func (a *app) tag(commit plumbing.Hash) error {
	return a.git.Tag(a.version, a.updateMajorMinor, commit)
}

func (a *app) push() error {
	if err := a.git.Push(); err != nil {
		return utils.Wrap("error pushing commit: %s", err)
	}

	if err := a.git.PushTags(a.version, a.updateMajorMinor); err != nil {
		return utils.Wrap("error pushing tags: %s", err)
	}

	return nil
}
