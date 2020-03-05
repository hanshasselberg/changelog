package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

func signatureHelper() *object.Signature {
	when, _ := time.Parse(object.DateFormat, "Thu May 04 00:03:43 2017 +0200")
	return &object.Signature{
		Name:  "John Doe",
		Email: "john@doe.org",
		When:  when,
	}
}

func commitHelper(w *git.Worktree, msg string) (plumbing.Hash, error) {
	return w.Commit(msg, &git.CommitOptions{
		Author: signatureHelper(),
	})
}

func testRepo(t *testing.T) *git.Repository {
	r, err := git.Init(memory.NewStorage(), memfs.New())
	require.NoError(t, err)

	w, err := r.Worktree()
	require.NoError(t, err)
	var last plumbing.Hash
	last, err = commitHelper(w, `fix(agent): one one one`)
	require.NoError(t, err)
	last, err = commitHelper(w, `fix(dns): two two two`)
	require.NoError(t, err)
	_, err = r.CreateTag("v1.6.0", last, &git.CreateTagOptions{Tagger: signatureHelper(), Message: "v1.6.0"})
	require.NoError(t, err)
	ref := plumbing.NewReferenceFromStrings("refs/remotes/origin/release/1.6.x", last.String())
	err = r.Storer.SetReference(ref)
	last, err = commitHelper(w, `feat(dns): three three three`)
	require.NoError(t, err)
	last, err = commitHelper(w, `fix(dns): four four four`)
	require.NoError(t, err)
	_, err = r.CreateTag("v1.7.0", last, &git.CreateTagOptions{Tagger: signatureHelper(), Message: "v1.7.0"})
	last, err = commitHelper(w, `fix(dns): five five five`)
	_, err = r.CreateTag("v1.7.1", last, &git.CreateTagOptions{Tagger: signatureHelper(), Message: "v1.7.1"})
	last, err = commitHelper(w, `fix(dns): six six six`)
	require.NoError(t, err)
	ref = plumbing.NewReferenceFromStrings("refs/remotes/origin/release/1.7.x", last.String())
	err = r.Storer.SetReference(ref)
	require.NoError(t, err)

	return r
}

func TestIsReleaseBranch(t *testing.T) {
	b1 := ""
	require.False(t, isReleaseBranch(b1))
	b2 := "refs/remotes/origin/release/1.6.x"
	require.True(t, isReleaseBranch(b2))
	b3 := "refs/heads/master"
	require.True(t, isReleaseBranch(b3))
	b4 := "refs/remotes/origin/feature-something"
	require.False(t, isReleaseBranch(b4))
}

func TestBranches(t *testing.T) {
	r := testRepo(t)
	bs, err := branches(r)
	require.NoError(t, err)
	require.Len(t, bs, 3)
}

func TestCommits(t *testing.T) {
	r := testRepo(t)
	ref := plumbing.NewReferenceFromStrings("refs/heads/master", "")
	_, err := commits(r, ref)
	require.NoError(t, err)
}

func TestHashToRelease(t *testing.T) {
	r := testRepo(t)
	hashToReleaseMap, err := hashToRelease(r)
	require.NoError(t, err)
	require.Len(t, hashToReleaseMap, 3)

	ref := plumbing.NewReferenceFromStrings("refs/heads/master", "")
	cms, err := commits(r, ref)
	// t.Log(hashToReleaseMap)
	// t.Log(cms)
	release, ok := hashToReleaseMap[cms[4].Hash.String()]
	require.True(t, ok)
	require.Equal(t, "1.6.0", release)

	release, ok = hashToReleaseMap[cms[2].Hash.String()]
	require.True(t, ok)
	require.Equal(t, "1.7.0", release)

	release, ok = hashToReleaseMap[cms[1].Hash.String()]
	require.True(t, ok)
	require.Equal(t, "1.7.1", release)

}

func TestChangelogChangelog(t *testing.T) {
	r := testRepo(t)
	ref := plumbing.NewReferenceFromStrings("refs/heads/master", "")
	c, err := newChangelog(r, ref)
	require.NoError(t, err)
	md, err := c.Changelog()
	require.NoError(t, err)
	t.Log(md)
}
