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
	hashToReleaseMap, err := hashToRelease(r)
	if err != nil {
		fmt.Println("error hashToRelease:", err)
		os.Exit(1)
	}

	ref := plumbing.NewReferenceFromStrings("refs/heads/master", "")
	commits, err := commits(r, ref)
	if err != nil {
		fmt.Println("error commits:", err)
		os.Exit(1)
	}
	md, err := changelog(hashToReleaseMap, commits)
	if err != nil {
		fmt.Println("error changelog:", err)
		os.Exit(1)
	}
	fmt.Println(md)
}

var releaseBranchRegex = regexp.MustCompile(`^refs\/remotes\/\w+\/release\/\d+\.\d+\.x$`)

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

var releaseTagRegex = regexp.MustCompile(`^v(\d+\.\d+\.\d+)$`)

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

var commitRegex = regexp.MustCompile(`^(fix|feat|impr|sec|note)\((\w+)\)!?: (.+)`)

func validCommit(msg string) bool {
	return commitRegex.MatchString(msg)
}

func formatCommit(msg string) (string, string) {
	matches := commitRegex.FindStringSubmatch(msg)
	cat := ""
	switch matches[1] {
	case "fix":
		cat = "BUGFIX"
	case "feat":
		cat = "FEATURE"
	case "impr":
		cat = "IMPROVEMENT"
	case "sec":
		cat = "SECURITY"
	case "note":
		cat = "NOTE"
	}
	return cat, fmt.Sprintf("* %s: %s", matches[2], matches[3])
}

func changelog(hashToReleaseMap map[string]string, commits []*object.Commit) (string, error) {
	currRelease := "UNRELEASED"
	releasesMap := map[string][]*object.Commit{currRelease: []*object.Commit{}}
	for _, cm := range commits {
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
		result = append(result, fmt.Sprintf("\n## %s", r))
		release := map[string][]string{}
		for _, cm := range releasesMap[r] {
			if !validCommit(cm.Message) {
				continue
			}
			cat, msg := formatCommit(cm.Message)
			release[cat] = append(release[cat], msg)
		}
		for _, cat := range []string{"SECURITY", "FEATURE", "IMPROVEMENT", "BUGFIX", "NOTE"} {
			if cms, ok := release[cat]; ok {
				result = append(result, fmt.Sprintf("\n%s\n", cat))
				result = append(result, sort.StringSlice(cms)...)
			}
		}
	}
	return strings.Join(result, "\n"), nil
}
