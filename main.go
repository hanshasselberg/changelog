package main

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

type changelog struct {
	branches map[string]string
}

func newChangelog() *changelog {
	return &changelog{branches: map[string]string{}}
}

func (c *changelog) add(branch, text string) {
	c.branches[branch] = text
}

func main() {
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: "../changelogrepo",
	})
	if err != nil {
		fmt.Println("error cloning:", err)
		os.Exit(1)
	}
	_, err = branches(r)
	if err != nil {
		fmt.Println("error branches:", err)
		os.Exit(1)
	}

	// c := newChangelog()
	// for _, b := range branches {
	// 	c.add(b.Name().String(), "")
	// }
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

func branches(r *git.Repository) ([]*plumbing.Reference, error) {
	var branches []*plumbing.Reference
	refs, err := r.References()
	if err != nil {
		return nil, err
	}
	refs.ForEach(func(ref *plumbing.Reference) error {
		if isReleaseBranch(ref.Name().String()) {
			branches = append(branches, ref)
		}
		return nil
	})
	return branches, nil
}
