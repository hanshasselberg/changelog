package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

func testMaster(t *testing.T) *config.Branch {
	r := testRepo(t)
	b, err := r.Branch("master")
	require.NoError(t, err)
	return b
}

func testRepo(t *testing.T) *git.Repository {
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: "../changelogrepo",
	})
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

func TestChangelogFor(t *testing.T) {
}
