package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func main() {
	path := flag.String("path", ".", "path to repository")
	flag.Parse()
	r, err := git.PlainOpen(*path)
	if err != nil {
		fmt.Println("error opening:", err)
		os.Exit(1)
	}
	ref := plumbing.NewReferenceFromStrings("refs/heads/master", "")
	c, err := newChangelog(r, ref)
	if err != nil {
		fmt.Println("error new changelog:", err)
		os.Exit(1)
	}
	md, err := c.Changelog()
	if err != nil {
		fmt.Println("error changelog:", err)
		os.Exit(1)
	}
	fmt.Println(md)
}

var releaseBranchRegex = regexp.MustCompile(`^refs\/remotes\/\w+\/release\/\d+\.\d+\.x$`)
var releaseTagRegex = regexp.MustCompile(`^v(\d+\.\d+\.\d+)$`)

func isReleaseBranch(b string) bool {
	if b == "refs/heads/master" {
		return true
	}
	if releaseBranchRegex.MatchString(b) {
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
		if isReleaseBranch(ref.Name().String()) {
			branches = append(branches, ref)
		}
		return nil
	})
	return branches, nil
}

type changelog struct {
	repository *git.Repository
	commits    []*object.Commit
}

func newChangelog(r *git.Repository, b *plumbing.Reference) (*changelog, error) {
	commits, err := commits(r, b)
	if err != nil {
		return nil, err
	}
	return &changelog{repository: r, commits: commits}, nil
}

func hashToRelease(r *git.Repository) (map[string]string, error) {
	ts, err := r.TagObjects()
	if err != nil {
		return nil, err
	}
	tagsMap := map[string]string{}
	ts.ForEach(func(t *object.Tag) error {
		if v := releaseTagRegex.FindStringSubmatch(t.Name); len(v) > 0 {
			tagsMap[t.Target.String()] = v[1]
		}
		return nil
	})
	return tagsMap, nil
}

func (c *changelog) Changelog() (string, error) {
	hashToReleaseMap, err := hashToRelease(c.repository)
	if err != nil {
		return "", err
	}
	currRelease := "UNRELEASED"
	releasesMap := map[string][]*object.Commit{currRelease: []*object.Commit{}}
	for _, cm := range c.commits {
		if r, ok := hashToReleaseMap[cm.Hash.String()]; ok {
			currRelease = r
			releasesMap[r] = []*object.Commit{}
		}
		releasesMap[currRelease] = append(releasesMap[currRelease], cm)
	}
	var releases sort.StringSlice
	for r := range releasesMap {
		releases = append(releases, r)
	}
	sort.Sort(sort.Reverse(releases))
	result := []string{}
	for _, r := range releases {
		result = append(result, fmt.Sprintf("\n## %s\n", r))
		for _, cm := range releasesMap[r] {
			result = append(result, fmt.Sprintf("* %s", strings.TrimSpace(cm.Message)))
		}
	}
	return strings.Join(result, "\n"), nil
}
