package main

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

type changelog struct {
	branches []*plumbing.Reference
}

func newChangelog() *changelog {
	return &changelog{branches: []*plumbing.Reference{}}
}

func (c *changelog) add(r *plumbing.Reference) {
	c.branches = append(c.branches, r)
}

func main() {
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: "../changelogrepo",
	})
	if err != nil {
		fmt.Println("error cloning:", err)
		os.Exit(1)
	}
	bs, err := branches(r)
	if err != nil {
		fmt.Println("error branches:", err)
		os.Exit(1)
	}

	c := changelog{}
	for _, b := range bs {
		c.add(b)
	}
}

var releaseRegex = regexp.MustCompile(`refs\/remotes\/\w+\/release\/`) // panic if regexp invalid

func isReleaseBranch(b string) bool {
	if b == "refs/heads/master" {
		return true
	}
	if releaseRegex.MatchString(b) {
		return true
	}
	return false
}

func commits(r *git.Repository, ref *plumbing.Reference) ([]*object.Commit, error) {
	var result []*object.Commit
	commits, err := r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, err
	}
	commits.ForEach(func(c *object.Commit) error {
		result = append(result, c)
		return nil
	})
	return result, nil
}

func branches(r *git.Repository) ([]*plumbing.Reference, error) {
	var branches []*plumbing.Reference
	refs, err := r.References()
	if err != nil {
		return nil, err
	}
	refs.ForEach(func(ref *plumbing.Reference) error {
		fmt.Println("ref", ref.Name().String(), ref.Name().IsBranch())
		if isReleaseBranch(ref.Name().String()) {
			branches = append(branches, ref)
		}
		return nil
	})
	return branches, nil
}
