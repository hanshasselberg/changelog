package main

import (
	"strings"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/require"
)

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

func TestCommits(t *testing.T) {
	r, err := testRepo()
	require.NoError(t, err)
	ref := plumbing.NewReferenceFromStrings("refs/heads/master", "")
	_, err = commits(r, ref)
	require.NoError(t, err)
}

func TestHashToRelease(t *testing.T) {
	r, err := testRepo()
	require.NoError(t, err)
	hashToReleaseMap, err := hashToRelease(r)
	require.NoError(t, err)
	require.Len(t, hashToReleaseMap, 1)

	ref := plumbing.NewReferenceFromStrings("refs/heads/master", "")
	cms, err := commits(r, ref)
	release, ok := hashToReleaseMap[cms[3].Hash.String()]
	require.True(t, ok)
	require.Equal(t, "1.7.3", release)

}

func TestChangelog(t *testing.T) {
	r, err := testRepo()
	require.NoError(t, err)
	ref := plumbing.NewReferenceFromStrings("refs/heads/master", "")
	commits, err := commits(r, ref)
	require.NoError(t, err)
	hashToRelease, err := hashToRelease(r)
	require.NoError(t, err)
	md, err := changelog(hashToRelease, commits)
	require.NoError(t, err)
	t.Log(md)
}

func TestValidHeadline(t *testing.T) {
	r, err := testRepo()
	require.NoError(t, err)
	w, err := r.Worktree()
	require.NoError(t, err)
	hash, err := commitHelper(w, 7777, "fubar")
	require.NoError(t, err)
	c, err := r.CommitObject(hash)
	require.NoError(t, err)
	headline := strings.Split(c.Message, "\n")[0]
	require.True(t, validHeadline(headline), headline)
}

func TestValidEntry(t *testing.T) {
	require.True(t, validEntry("* fix(agent): something something"))
	require.True(t, validEntry("* feat(agent): something something"))
	require.True(t, validEntry("* impr(agent): something something"))
	require.True(t, validEntry("* sec(agent): something something"))
	require.True(t, validEntry("* note(agent): something something"))
	require.True(t, validEntry("* feat(agent)!: something something"))
	require.True(t, validEntry("* feat(agent):something something"))
	require.True(t, validEntry("* feat(__agent fuu__):something something"))
	require.False(t, validEntry("* wat(agent): something something"))
	require.False(t, validEntry("* aeoustaouesth"))
}

func TestExtractCatAndChange(t *testing.T) {
	cat, change := extractCatAndChange("fix(agent): something")
	require.Equal(t, "BUGFIX", cat)
	require.Equal(t, "* agent: something", change)
	cat, change = extractCatAndChange("fix(agent)!: something")
	require.Equal(t, "BUGFIX", cat)
	require.Equal(t, "* agent: something", change)
}

func TestEntriesFromMessage(t *testing.T) {
	msg := "something #123\nfubar ok\n```changelog\n* fix(dns): first\n* feat(agent): second"
	entries := entriesFromMessage(msg)
	require.Len(t, entries, 2)
	require.Equal(t, "* fix(dns): first", entries[0])
	require.Equal(t, "* feat(agent): second", entries[1])
}
