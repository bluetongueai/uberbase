package deploy

import (
	"testing"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestGitCloner(t *testing.T) {
	mock := core.NewMockSSH("localhost:2222")
	go mock.ListenAndServe()
	defer mock.Close()

	t.Run("Clone Repository", func(t *testing.T) {
		conn, err := core.NewSession(core.SSHConfig{
			Host:    "localhost:2222",
			User:    "test-user",
			KeyData: "dummy-key",
		})
		assert.NoError(t, err)
		defer conn.Close()

		cloner := NewGitCloner(conn)

		opts := CloneOptions{
			URL:         "git@github.com:user/repo.git",
			Branch:      "main",
			Destination: "/app",
			Depth:       1,
		}

		mock.SetReturnString("")

		err = cloner.Clone(opts)
		assert.NoError(t, err)
	})

	t.Run("Get Remote URL", func(t *testing.T) {
		conn, err := core.NewSession(core.SSHConfig{
			Host:    "localhost:2222",
			User:    "test-user",
			KeyData: "dummy-key",
		})
		assert.NoError(t, err)
		defer conn.Close()

		cloner := NewGitCloner(conn)

		expectedURL := "git@github.com:user/repo.git"
		mock.SetReturnString(expectedURL)

		url, err := cloner.GetRemoteURL("/app")
		assert.NoError(t, err)
		assert.Equal(t, expectedURL, url)
	})

	t.Run("Git Not Installed", func(t *testing.T) {
		conn, err := core.NewSession(core.SSHConfig{
			Host:    "localhost:2222",
			User:    "test-user",
			KeyData: "dummy-key",
		})
		assert.NoError(t, err)
		defer conn.Close()

		cloner := NewGitCloner(conn)

		mock.SetReturnString("git not found")

		opts := CloneOptions{URL: "git@github.com:user/repo.git"}
		err = cloner.Clone(opts)
		assert.Error(t, err)
	})

	t.Run("Is Git Repository", func(t *testing.T) {
		conn, err := core.NewSession(core.SSHConfig{
			Host:    "localhost:2222",
			User:    "test-user",
			KeyData: "dummy-key",
		})
		assert.NoError(t, err)
		defer conn.Close()

		cloner := NewGitCloner(conn)

		mock.SetReturnString("")
		assert.True(t, cloner.IsGitRepository("/app"))

		mock.SetReturnString("not a git repo")
		assert.False(t, cloner.IsGitRepository("/not-git"))
	})

	t.Run("Get Current Commit", func(t *testing.T) {
		conn, err := core.NewSession(core.SSHConfig{
			Host:    "localhost:2222",
			User:    "test-user",
			KeyData: "dummy-key",
		})
		assert.NoError(t, err)
		defer conn.Close()

		cloner := NewGitCloner(conn)

		expectedCommit := "abc123"
		mock.SetReturnString(expectedCommit)

		commit, err := cloner.GetCurrentCommit("/app")
		assert.NoError(t, err)
		assert.Equal(t, expectedCommit, commit)
	})
}

func TestGitCloner_Update(t *testing.T) {
	mock := core.NewMockSSH("localhost:2222")
	go mock.ListenAndServe()
	defer mock.Close()

	conn, err := core.NewSession(core.SSHConfig{
		Host:    "localhost:2222",
		User:    "test-user",
		KeyData: "dummy-key",
	})
	assert.NoError(t, err)
	defer conn.Close()

	cloner := NewGitCloner(conn)

	t.Run("Successful Update", func(t *testing.T) {
		mock.SetReturnString("")

		err := cloner.Update("/app", "main")
		assert.NoError(t, err)
	})

	t.Run("Update with Fetch Error", func(t *testing.T) {
		mock.SetReturnString("network error")

		err := cloner.Update("/app", "main")
		assert.Error(t, err)
	})
}

func TestGitCloner_EdgeCases(t *testing.T) {
	mock := core.NewMockSSH("localhost:2222")
	go mock.ListenAndServe()
	defer mock.Close()

	conn, err := core.NewSession(core.SSHConfig{
		Host:    "localhost:2222",
		User:    "test-user",
		KeyData: "dummy-key",
	})
	assert.NoError(t, err)
	defer conn.Close()

	cloner := NewGitCloner(conn)

	t.Run("Clone with Special Characters", func(t *testing.T) {
		mock.SetReturnString("")

		opts := CloneOptions{
			URL:         "git@github.com:user/repo with spaces.git",
			Branch:      "feature/branch-with-slash",
			Destination: "/path/with spaces/and$special&chars",
		}

		err := cloner.Clone(opts)
		assert.NoError(t, err)
	})

	t.Run("Empty URL", func(t *testing.T) {
		opts := CloneOptions{
			URL: "",
		}
		err := cloner.Clone(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "URL is required")
	})
}
