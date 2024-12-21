package pkg

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGitCloner(t *testing.T) {
	t.Run("Clone Repository", func(t *testing.T) {
			mock := NewMockSSHClient()
			cloner := NewGitCloner(mock)

			opts := CloneOptions{
				URL:         "git@github.com:user/repo.git",
				Branch:      "main",
				Destination: "/app",
				Depth:       1,
			}

			mock.SetOutput("which git", "")
			mock.SetOutput("mkdir -p /app", "")
			mock.SetOutput("git -c advice.detachedHead=false clone", "")

			if err := cloner.Clone(opts); err != nil {
				t.Errorf("Clone failed: %v", err)
			}

			commands := mock.GetCommands()
			lastCmd := commands[len(commands)-1]
			expectedCmd := "git -c advice.detachedHead=false clone --porcelain -b main --depth=1 git@github.com:user/repo.git /app"
			if lastCmd != expectedCmd {
				t.Errorf("Expected command %q, got %q", expectedCmd, lastCmd)
			}
	})

	t.Run("Get Remote URL", func(t *testing.T) {
		mock := NewMockSSHClient()
		cloner := NewGitCloner(mock)

		expectedURL := "git@github.com:user/repo.git"
		mock.SetOutput("cd /app && git config --get remote.origin.url", expectedURL)

		url, err := cloner.GetRemoteURL("/app")
		if err != nil {
			t.Errorf("GetRemoteURL failed: %v", err)
		}
		if url != expectedURL {
			t.Errorf("Expected URL %q, got %q", expectedURL, url)
		}
	})

	t.Run("Git Not Installed", func(t *testing.T) {
		mock := NewMockSSHClient()
		cloner := NewGitCloner(mock)

		mock.SetError("which git", fmt.Errorf("git not found"))

		opts := CloneOptions{URL: "git@github.com:user/repo.git"}
		if err := cloner.Clone(opts); err == nil {
			t.Error("Expected error when git is not installed")
		}
	})

	t.Run("Is Git Repository", func(t *testing.T) {
		mock := NewMockSSHClient()
		cloner := NewGitCloner(mock)

		mock.SetOutput("test -d /app/.git", "")
		if !cloner.IsGitRepository("/app") {
			t.Error("Expected /app to be a git repository")
		}

		mock = NewMockSSHClient()
		cloner = NewGitCloner(mock)
		mock.SetError("test -d /not-git/.git", fmt.Errorf("not a git repo"))
		if cloner.IsGitRepository("/not-git") {
			t.Error("Expected /not-git to not be a git repository")
		}
	})

	t.Run("Get Current Commit", func(t *testing.T) {
		mock := NewMockSSHClient()
		cloner := NewGitCloner(mock)

		expectedCommit := "abc123"
		mock.SetOutput("cd /app && git rev-parse HEAD", expectedCommit)

		commit, err := cloner.GetCurrentCommit("/app")
		if err != nil {
			t.Errorf("GetCurrentCommit failed: %v", err)
		}
		if commit != expectedCommit {
			t.Errorf("Expected commit %q, got %q", expectedCommit, commit)
		}
	})

	t.Run("Clone Error Handling", func(t *testing.T) {
		mock := NewMockSSHClient()
		cloner := NewGitCloner(mock)

		mock.SetOutput("which git", "")
		mock.SetOutput("mkdir -p /app", "")
		mock.SetError("git -c advice.detachedHead=false clone", fmt.Errorf("clone failed"))

		opts := CloneOptions{
			URL:         "git@github.com:user/repo.git",
			Destination: "/app",
		}
		if err := cloner.Clone(opts); err == nil {
			t.Error("Expected error when clone fails")
		}
	})
}

func TestGitCloner_Update(t *testing.T) {
	mock := NewMockSSHClient()
	cloner := NewGitCloner(mock)

	t.Run("Successful Update", func(t *testing.T) {
		mock.SetOutput("git -c advice.detachedHead=false fetch", "")
		mock.SetOutput("git -c advice.detachedHead=false reset", "")
		mock.SetOutput("git clean", "")

		err := cloner.Update("/app", "main")
		if err != nil {
			t.Errorf("Update failed: %v", err)
		}

		commands := mock.GetCommands()
		expectedCmds := []string{
			"cd /app && git -c advice.detachedHead=false fetch --porcelain origin main",
			"cd /app && git -c advice.detachedHead=false reset --hard origin/main",
			"cd /app && git clean -fd",
		}

		for i, expected := range expectedCmds {
			if !strings.Contains(commands[i], expected) {
				t.Errorf("Expected command %q, got %q", expected, commands[i])
			}
		}
	})

	t.Run("Update with Fetch Error", func(t *testing.T) {
		mock.SetError("git -c advice.detachedHead=false fetch", fmt.Errorf("network error"))
		
		err := cloner.Update("/app", "main")
		if err == nil {
			t.Error("Expected error on fetch failure")
		}
	})
}

func TestGitCloner_EdgeCases(t *testing.T) {
	mock := NewMockSSHClient()
	cloner := NewGitCloner(mock)

	t.Run("Clone with Special Characters", func(t *testing.T) {
		mock.SetOutput("which git", "")
		mock.SetOutput("mkdir -p", "")
		mock.SetOutput("git -c advice.detachedHead=false clone", "")

		opts := CloneOptions{
			URL:         "git@github.com:user/repo with spaces.git",
			Branch:      "feature/branch-with-slash",
			Destination: "/path/with spaces/and$special&chars",
		}

		err := cloner.Clone(opts)
		assert.NoError(t, err)

		// Verify command escaping
		commands := mock.GetCommands()
		lastCmd := commands[len(commands)-1]
		assert.Contains(t, lastCmd, `"/path/with spaces/and$special&chars"`)
		assert.Contains(t, lastCmd, `"git@github.com:user/repo with spaces.git"`)
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

func TestGitCloner_Concurrent(t *testing.T) {
	t.Run("Concurrent Clones", func(t *testing.T) {
		mock := NewMockSSHClient()
		cloner := NewGitCloner(mock)

		// Set up common mock responses
		mock.SetOutput("which git", "")
		mock.SetOutput("git -c advice.detachedHead=false clone", "")
		mock.SetOutput("mkdir -p", "") // Generic response for all mkdir commands

		var wg sync.WaitGroup
		errs := make([]error, 0)
		var errMu sync.Mutex

		// Start goroutines
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				
				opts := CloneOptions{
					URL:         fmt.Sprintf("git@github.com:user/repo%d.git", id),
					Destination: fmt.Sprintf("/app%d", id),
				}

				if err := cloner.Clone(opts); err != nil {
					errMu.Lock()
					errs = append(errs, err)
					errMu.Unlock()
				}
			}(i)
		}

		// Wait with timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			if len(errs) > 0 {
				t.Errorf("Got %d clone errors: %v", len(errs), errs)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Test timed out")
		}

		// Verify commands were executed
		commands := mock.GetCommands()
		assert.GreaterOrEqual(t, len(commands), 5, 
			"Expected at least 5 commands, got %d", len(commands))
	})
}

func TestGitCloner_NetworkFailures(t *testing.T) {
	mock := NewMockSSHClient()
	cloner := NewGitCloner(mock)

	t.Run("Network Timeout", func(t *testing.T) {
		mock.SetOutput("which git", "")
		mock.SetOutput("mkdir", "")
		mock.SetError("git -c advice.detachedHead=false clone", fmt.Errorf("timeout"))

		opts := CloneOptions{
			URL:         "git@github.com:user/repo.git",
			Destination: "/app",
		}

		err := cloner.Clone(opts)
		if err == nil {
			t.Error("Expected error on network timeout")
		}
	})

	t.Run("SSH Authentication Failure", func(t *testing.T) {
		mock.SetOutput("which git", "")
		mock.SetOutput("mkdir", "")
		mock.SetError("git -c advice.detachedHead=false clone", fmt.Errorf("permission denied (publickey)"))

		opts := CloneOptions{
			URL:         "git@github.com:user/repo.git",
			Destination: "/app",
		}

		err := cloner.Clone(opts)
		if err == nil {
			t.Error("Expected error on SSH authentication failure")
		}
	})
}

func TestGitCloner_Validation(t *testing.T) {
	mock := NewMockSSHClient()
	cloner := NewGitCloner(mock)

	tests := []struct {
		name    string
		opts    CloneOptions
		wantErr bool
		errMsg  string
	}{
		{
			name: "invalid URL format",
			opts: CloneOptions{
				URL: "not-a-git-url",
			},
			wantErr: true,
			errMsg:  "invalid git URL format",
		},
		{
			name: "invalid branch characters",
			opts: CloneOptions{
				URL:    "git@github.com:user/repo.git",
					Branch: "branch with spaces",
			},
			wantErr: true,
			errMsg:  "invalid branch name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cloner.Clone(tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
} 
