package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

func main() {
	path := flag.String("path", "", "path to repository")
	flag.Parse()
	var r *git.Repository
	var err error
	if *path != "" {
		r, err = git.PlainOpen(*path)
		if err != nil {
			fmt.Println("error opening:", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Using test repo\n\n")
		r, err = testRepo()
		if err != nil {
			fmt.Println("error opening:", err)
			os.Exit(1)
		}
	}
	hashToReleaseMap, err := hashToRelease(r)
	if err != nil {
		fmt.Println("error hashToRelease:", err)
		os.Exit(1)
	}

	ref := plumbing.NewReferenceFromStrings("HEAD", "")
	// ref := plumbing.NewReferenceFromStrings("refs/heads/master", "")
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

	if *path == "" {
		fmt.Printf("commits: \n\n")
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 1, 0, 1, ' ', 0)
		for _, c := range commits {
			fmt.Fprintf(w, "%s\t%s\t%s\n", c.Hash.String()[0:6], hashToReleaseMap[c.Hash.String()], strings.Replace(c.Message, "\n", " ", -1))
		}
		w.Flush()
		fmt.Printf("\nchangelog: \n")
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

var headlineRegex = regexp.MustCompile(`#\d{4,5}`)
var entryRegex = regexp.MustCompile(`^\*? ?(fix|feat|impr|sec|note)\((\w+)\)!?:(.*)(\(#\d+\))?`)

func validHeadline(line string) bool {
	return headlineRegex.MatchString(line)
}

func validEntry(entry string) bool {
	return entryRegex.MatchString(entry)
}

func entriesFromMessage(msg string) []string {
	entries := []string{}
	var found bool
	for _, l := range strings.Split(msg, "\n") {
		l = strings.TrimSpace(l)
		if l == "```changelog" {
			found = true
			continue
		} else if l == "```" {
			found = false
			continue
		}
		if found {
			entries = append(entries, l)
		}
	}
	return entries
}

func extractCatAndChange(entry string) (string, string) {
	matches := entryRegex.FindStringSubmatch(entry)
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
	return cat, fmt.Sprintf("* %s: %s", strings.TrimSpace(matches[2]), strings.TrimSpace(matches[3]))
}

func changelog(hashToReleaseMap map[string]string, commits []*object.Commit) (string, error) {
	currRelease := []*object.Commit{}
	for _, cm := range commits {
		if _, ok := hashToReleaseMap[cm.Hash.String()]; ok {
			break
		}
		currRelease = append(currRelease, cm)
	}
	result := []string{}
	release := map[string][]string{}
	for _, cm := range currRelease {
		if !validHeadline(cm.Message) {
			continue
		}
		headline := strings.Split(cm.Message, "\n")[0]
		if validHeadline(headline) {
			for _, e := range entriesFromMessage(cm.Message) {
				if !validEntry(e) {
					continue
				}
				cat, change := extractCatAndChange(e)
				release[cat] = append(release[cat], change)
			}
		}
	}
	for _, cat := range []string{"SECURITY", "FEATURE", "IMPROVEMENT", "BUGFIX", "NOTE"} {
		if cms, ok := release[cat]; ok {
			result = append(result, fmt.Sprintf("\n%s\n", cat))
			result = append(result, sort.StringSlice(cms)...)
		}
	}
	return strings.Join(result, "\n"), nil
}

func signatureHelper() *object.Signature {
	when, _ := time.Parse(object.DateFormat, "Thu May 04 00:03:43 2017 +0200")
	return &object.Signature{
		Name:  "John Doe",
		Email: "john@doe.org",
		When:  when,
	}
}

func commitHelper(w *git.Worktree, pr int, entry string) (plumbing.Hash, error) {
	msg := fmt.Sprintf("Merge pull request #%d from hashicorp/something\nother notes\n```changelog\n%s\n```\n", pr, entry)
	return w.Commit(msg, &git.CommitOptions{
		Author: signatureHelper(),
	})
}

func testRepo() (*git.Repository, error) {
	r, err := git.Init(memory.NewStorage(), memfs.New())
	if err != nil {
		return nil, err
	}

	w, err := r.Worktree()
	if err != nil {
		return nil, err
	}
	last, err := commitHelper(w, 5555, `fix(dns): five five five`)
	if err != nil {
		return nil, err
	}
	_, err = r.CreateTag("v1.7.3", last, &git.CreateTagOptions{Tagger: signatureHelper(), Message: "v1.7.1"})
	if err != nil {
		return nil, err
	}
	_, err = commitHelper(w, 6666, `fix(dns): six six six`)
	if err != nil {
		return nil, err
	}
	_, err = commitHelper(w, 7777, `feat(agent): seven seven seven`)
	if err != nil {
		return nil, err
	}
	_, err = commitHelper(w, 8888, `feat(agent): eight eight eight`)
	if err != nil {
		return nil, err
	}

	return r, nil
}
