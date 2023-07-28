package repository

import (
	"fmt"
	"time"

	"changelog-version/repository/api"
	"changelog-version/utils"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"

	"golang.org/x/mod/semver"
)

type Repository struct {
	committer *object.Signature
	author    *object.Signature
	auth      *githttp.BasicAuth
	repo      *git.Repository
	worktree  *git.Worktree
}

func New(token, actor string) (*Repository, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return nil, utils.Wrap("error opening repository: %s", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return nil, utils.Wrap("error identifying worktree: %s", err)
	}

	tm := time.Now()
	committer := &object.Signature{
		Name:  "github-actions[bot]",
		Email: "github-actions[bot]@users.noreply.github.com",
		When:  tm,
	}

	author := &object.Signature{
		Name: actor,
		When: tm,
	}

	if author.Name == "" {
		author = committer
	} else {
		a := api.New(token)
		email, err := a.GetUserEmail(author.Name)
		if err != nil {
			return nil, utils.Wrap("error fetching author email: %s", err)
		}

		if email != "" {
			author.Email = email
		}
	}

	return &Repository{
		committer: committer,
		author:    author,
		auth: &githttp.BasicAuth{
			Username: committer.Name,
			Password: token,
		},
		repo:     r,
		worktree: w,
	}, nil
}

func (r *Repository) isTagExists(name string) (bool, error) {
	tags, err := r.repo.TagObjects()
	if err != nil {
		return false, utils.Wrap("error fetching tags: %s", err)
	}

	res := false
	err = tags.ForEach(func(t *object.Tag) error {
		if t.Name == name {
			res = true
			return git.ErrTagExists
		}
		return nil
	})
	if err != nil && err != git.ErrTagExists {
		return false, utils.Wrap("tags iterator error: %s", err)
	}
	return res, nil
}

func getRefs(version string, updateMajorMinor bool) []string {
	refs := []string{version}
	if updateMajorMinor {
		refs = append(refs, []string{
			semver.Major(version),
			semver.MajorMinor(version),
		}...)
	}

	return refs
}

func (r *Repository) Commit(message string) (plumbing.Hash, error) {
	if _, err := r.worktree.Add("."); err != nil {
		return plumbing.Hash{}, utils.Wrap("error staging updated files: %s", err)
	}

	c, err := r.worktree.Commit(
		message,
		&git.CommitOptions{
			Author:    r.author,
			Committer: r.committer,
		},
	)
	if err != nil {
		return plumbing.Hash{}, utils.Wrap("error committing changes: %s", err)
	}

	return c, nil
}

func (r *Repository) Tag(name string, updateMajorMinor bool, commit plumbing.Hash) error {
	exists, err := r.isTagExists(name)
	if err != nil {
		return utils.Wrap("error validating tag existence: %s", err)
	}
	if exists {
		return git.ErrTagExists
	}

	h, err := r.repo.Head()
	if err != nil {
		return utils.Wrap("error identifying head reference: %s", err)
	}

	for _, v := range getRefs(name, updateMajorMinor) {
		_, err = r.repo.CreateTag(v, h.Hash(), &git.CreateTagOptions{
			Message: v,
			Tagger:  r.committer,
		})
		if err != nil {
			if v != name && err == git.ErrTagExists {
				if err := r.repo.DeleteTag(v); err != nil {
					return utils.Wrap("error deleting tag: %s", err)
				}

				if _, err = r.repo.CreateTag(v, h.Hash(), &git.CreateTagOptions{
					Message: v,
					Tagger:  r.committer,
				}); err != nil {
					return utils.Wrap("error tagging a commit (%s): %s", v, err)
				}

				continue
			}

			return utils.Wrap("error tagging a commit (%s): %s", name, err)
		}
	}

	return nil
}

func (r *Repository) Push() error {
	return r.repo.Push(&git.PushOptions{Auth: r.auth})
}

func (r *Repository) PushTags(name string, updateMajorMinor bool) error {
	for _, v := range getRefs(name, updateMajorMinor) {
		o := &git.PushOptions{
			RefSpecs: []config.RefSpec{config.RefSpec(
				fmt.Sprintf("refs/tags/%s:refs/tags/%s", v, v),
			)},
			Auth: r.auth,
		}

		if v != name {
			o.Force = true
		}

		err := r.repo.Push(o)
		if err != nil {
			return utils.Wrap("error pushing the tag (%s): %s", v, err)
		}
	}

	return nil
}
