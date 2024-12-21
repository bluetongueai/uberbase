package pkg

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilder(t *testing.T) {
	t.Run("basic build", func(t *testing.T) {
		ssh := NewMockSSHClient()
		builder := NewBuilder(ssh)

		opts := BuildOptions{
			File:        "Dockerfile",
			ContextPath: ".",
			Tag:        "myapp:latest",
		}

		// Set up mock responses
		ssh.SetOutput("test -d .", "")
		ssh.SetOutput("test -f ./Dockerfile", "")
		ssh.SetOutput("podman build", "")

		err := builder.Build(opts)
		assert.NoError(t, err)

		// Get all commands and find the build command
		commands := ssh.GetCommands()
		var buildCmd string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, "podman build") {
				buildCmd = cmd
				break
			}
		}
		
		// Verify the build command contains expected flags
		assert.NotEmpty(t, buildCmd, "Build command not found in executed commands")
		assert.Contains(t, buildCmd, "-f Dockerfile")
		assert.Contains(t, buildCmd, "-t myapp:latest")
		assert.Contains(t, buildCmd, ".")
	})

	t.Run("build with args", func(t *testing.T) {
		ssh := NewMockSSHClient()
		builder := NewBuilder(ssh)

		opts := BuildOptions{
			File:        "Dockerfile",
			ContextPath: ".",
			Tag:        "test:latest",
			BuildArgs: map[string]string{
				"VERSION": "1.0",
				"DEBUG":   "true",
			},
		}

		// Set up mock responses
		ssh.SetOutput("test -d .", "")
		ssh.SetOutput("test -f ./Dockerfile", "")
		ssh.SetOutput("podman build", "")

		err := builder.Build(opts)
		assert.NoError(t, err)

		// Get all commands and find the build command
		commands := ssh.GetCommands()
		var buildCmd string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, "podman build") {
				buildCmd = cmd
				break
			}
		}
		
		// Verify the build command contains expected flags
		assert.NotEmpty(t, buildCmd, "Build command not found in executed commands")
		assert.Contains(t, buildCmd, "--build-arg VERSION=1.0")
		assert.Contains(t, buildCmd, "--build-arg DEBUG=true")
	})

	t.Run("build with missing tag", func(t *testing.T) {
		ssh := NewMockSSHClient()
		builder := NewBuilder(ssh)

		opts := BuildOptions{
			File:        "Dockerfile",
			ContextPath: ".",
		}

		err := builder.Build(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tag is required")
	})

	t.Run("build with missing context", func(t *testing.T) {
		ssh := NewMockSSHClient()
		builder := NewBuilder(ssh)

		opts := BuildOptions{
			File: "Dockerfile",
			Tag:  "test:latest",
		}

		err := builder.Build(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context path is required")
	})

	t.Run("build with environment", func(t *testing.T) {
		ssh := NewMockSSHClient()
		builder := NewBuilder(ssh)

		opts := BuildOptions{
			File:        "Dockerfile",
			ContextPath: ".",
			Tag:        "test:latest",
			Environment: map[string]string{
				"BUILD_ENV": "test",
			},
		}

		// Set up mock responses
		ssh.SetOutput("test -d .", "")
		ssh.SetOutput("test -f ./Dockerfile", "")
		ssh.SetOutput("podman build", "")

		err := builder.Build(opts)
		assert.NoError(t, err)

		// Get all commands and find the build command
		commands := ssh.GetCommands()
		var buildCmd string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, "podman build") {
				buildCmd = cmd
				break
			}
		}
		
		// Verify the build command contains expected flags
		assert.NotEmpty(t, buildCmd, "Build command not found in executed commands")
		assert.Contains(t, buildCmd, "--env BUILD_ENV=test")
	})
}

func TestBuilder_Validation(t *testing.T) {
	tests := []struct {
		name    string
		opts    BuildOptions
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid options",
			opts: BuildOptions{
				File:        "Dockerfile",
				ContextPath: ".",
				Tag:        "test:latest",
			},
			wantErr: false,
		},
		{
			name: "invalid platform format",
			opts: BuildOptions{
				File:        "Dockerfile",
				ContextPath: ".",
				Tag:        "test:latest",
				Platform:   []string{"invalid"},
			},
			wantErr: true,
			errMsg:  "invalid platform format",
		},
		{
			name: "invalid pull value",
			opts: BuildOptions{
				File:        "Dockerfile",
				ContextPath: ".",
				Tag:        "test:latest",
				Pull:       "invalid",
			},
			wantErr: true,
			errMsg:  "invalid pull value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ssh := NewMockSSHClient()
			builder := NewBuilder(ssh)

			// Set up mock responses for valid cases
			if !tt.wantErr {
				ssh.SetOutput("test -d .", "")
				ssh.SetOutput("test -f ./Dockerfile", "")
				ssh.SetOutput("podman build", "")
			}

			err := builder.Build(tt.opts)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuilder_BuildWithEnv(t *testing.T) {
	mock := NewMockSSHClient()
	builder := NewBuilder(mock)

	t.Run("Build With Environment", func(t *testing.T) {
		// Set up mock responses
		mock.SetOutput("test -d .", "")
		mock.SetOutput("test -f ./Dockerfile", "")
		mock.SetOutput("podman build", "")

		opts := BuildOptions{
			File:        "Dockerfile",
			ContextPath: ".",
			Tag:        "test:latest",
			Environment: map[string]string{
				"BUILD_ENV": "test",
			},
		}

		err := builder.Build(opts)
		assert.NoError(t, err)

		commands := mock.GetCommands()
		var buildCmd string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, "podman build") {
				buildCmd = cmd
				break
			}
		}

		assert.NotEmpty(t, buildCmd, "Build command not found")
		assert.Contains(t, buildCmd, "--env BUILD_ENV=test")
	})

	t.Run("Build With Context Validation", func(t *testing.T) {
		mock.SetError("test -d /nonexistent", fmt.Errorf("directory not found"))

		opts := BuildOptions{
			File:        "Dockerfile",
			ContextPath: "/nonexistent",
			Tag:        "test:latest",
		}

		err := builder.Build(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context path does not exist")
	})
}

func TestBuilder_BuildCommandOptions(t *testing.T) {
	mock := NewMockSSHClient()
	builder := NewBuilder(mock)

	t.Run("Build With All Options", func(t *testing.T) {
		mock.SetOutput("podman build", "")
		opts := BuildOptions{
			File:        "custom.Dockerfile",
			ContextPath: ".",
			Tag:        "test:latest",
			BuildArgs:  map[string]string{"ARG1": "value1"},
			
			Target:     "prod",
			NoCache:    true,
			Pull:       "always",
			Platform:   []string{"linux/amd64"},
		}
		
		if err := builder.Build(opts); err != nil {
			t.Errorf("Build failed: %v", err)
		}

		commands := mock.GetCommands()
		lastCmd := commands[len(commands)-1]
		
		expectedFlags := []string{
			"-f custom.Dockerfile",
			"--build-arg ARG1=value1",
			"--target prod",
			"--no-cache",
			"--pull",
			"--platform linux/amd64",
		}

		for _, flag := range expectedFlags {
			if !strings.Contains(lastCmd, flag) {
				t.Errorf("Expected command to contain %q, got: %s", flag, lastCmd)
			}
		}
	})

	t.Run("Build Without Tag", func(t *testing.T) {
		opts := BuildOptions{
			File:        "Dockerfile",
			ContextPath: ".",
		}
		if err := builder.Build(opts); err == nil {
			t.Error("Expected error when tag is missing")
		}
	})

	t.Run("Build Without Context", func(t *testing.T) {
		opts := BuildOptions{
			File:        "Dockerfile",
			Tag:        "test:latest",
		}
		if err := builder.Build(opts); err == nil {
			t.Error("Expected error when context is missing")
		}
	})
}

func TestBuilder_ImageOperations(t *testing.T) {
	mock := NewMockSSHClient()
	builder := NewBuilder(mock)

	t.Run("List Images", func(t *testing.T) {
		expectedImages := "image1:latest\nimage2:v1"
		mock.SetOutput("podman images --format '{{.Repository}}:{{.Tag}}'", expectedImages)
		
		images, err := builder.ListImages()
		if err != nil {
			t.Errorf("ListImages failed: %v", err)
		}
		
		if len(images) != 2 {
			t.Errorf("Expected 2 images, got %d", len(images))
		}
		
		if images[0] != "image1:latest" || images[1] != "image2:v1" {
			t.Error("Unexpected image list content")
		}
	})

	t.Run("List Images Error", func(t *testing.T) {
		mock.SetError("podman images --format '{{.Repository}}:{{.Tag}}'", fmt.Errorf("command failed"))
		
		_, err := builder.ListImages()
		if err == nil {
			t.Error("Expected error when listing images fails")
		}
	})

	t.Run("Remove Image", func(t *testing.T) {
		mock.SetOutput("podman rmi", "")
		
		err := builder.RemoveImage("test:latest")
		if err != nil {
			t.Errorf("RemoveImage failed: %v", err)
		}
		
		commands := mock.GetCommands()
		lastCmd := commands[len(commands)-1]
		if !strings.Contains(lastCmd, "podman rmi test:latest") {
			t.Errorf("Unexpected command: %s", lastCmd)
		}
	})

	t.Run("Remove Image Error", func(t *testing.T) {
		mock.SetError("podman rmi", fmt.Errorf("command failed"))
		
		err := builder.RemoveImage("test:latest")
		if err == nil {
			t.Error("Expected error when removing image fails")
		}
	})

	t.Run("Prune Images", func(t *testing.T) {
		mock.SetOutput("podman image prune", "")
		
		err := builder.PruneImages()
		if err != nil {
			t.Errorf("PruneImages failed: %v", err)
		}
		
		commands := mock.GetCommands()
		lastCmd := commands[len(commands)-1]
		if !strings.Contains(lastCmd, "podman image prune -f") {
			t.Errorf("Unexpected command: %s", lastCmd)
		}
	})

	t.Run("Prune Images Error", func(t *testing.T) {
		mock.SetError("podman image prune", fmt.Errorf("command failed"))
		
		err := builder.PruneImages()
		if err == nil {
			t.Error("Expected error when pruning images fails")
		}
	})
}

func TestBuilder_ValidateContextPath(t *testing.T) {
	mock := NewMockSSHClient()
	builder := NewBuilder(mock)

	t.Run("Valid Context Path", func(t *testing.T) {
		mock.SetOutput("test -d", "")
		mock.SetOutput("test -f", "")
		
		err := builder.ValidateContextPath("/valid/path")
		if err != nil {
			t.Errorf("ValidateContextPath failed: %v", err)
		}
	})

	t.Run("Invalid Context Path", func(t *testing.T) {
		mock.SetError("test -d", fmt.Errorf("directory not found"))
		
		err := builder.ValidateContextPath("/invalid/path")
		if err == nil {
			t.Error("Expected error for invalid context path")
		}
	})

	t.Run("Missing Dockerfile", func(t *testing.T) {
		mock.SetOutput("test -d", "")
		mock.SetError("test -f", fmt.Errorf("file not found"))
		
		err := builder.ValidateContextPath("/path/without/dockerfile")
		if err == nil {
			t.Error("Expected error for missing Dockerfile")
		}
	})
}

func TestBuilder_BuildWithVolumes(t *testing.T) {
	mock := NewMockSSHClient()
	builder := NewBuilder(mock)

	t.Run("Build With Volumes", func(t *testing.T) {
		mock.SetOutput("podman build", "")
		opts := BuildOptions{
			File:        "Dockerfile",
			ContextPath: ".",
			Tag:        "test:latest",
			Volumes:    []string{"/host:/container:ro", "/data:/data"},
		}
		
		if err := builder.Build(opts); err != nil {
			t.Errorf("Build failed: %v", err)
		}

		commands := mock.GetCommands()
		lastCmd := commands[len(commands)-1]
		
		for _, volume := range opts.Volumes {
			if !strings.Contains(lastCmd, "--volume "+volume) {
				t.Errorf("Expected command to contain volume mount %q, got: %s", volume, lastCmd)
			}
		}
	})
}

func TestBuilder_BuildWithEnvironmentAndArgs(t *testing.T) {
	mock := NewMockSSHClient()
	builder := NewBuilder(mock)

	t.Run("Build With Build Args", func(t *testing.T) {
		mock.SetOutput("podman build", "")
		opts := BuildOptions{
			File:        "Dockerfile",
			ContextPath: ".",
			Tag:        "test:latest",
			BuildArgs: map[string]string{
				"ARG1": "value1",
			},
		}
		
		if err := builder.Build(opts); err != nil {
			t.Errorf("Build failed: %v", err)
		}

		commands := mock.GetCommands()
		lastCmd := commands[len(commands)-1]
		
		if !strings.Contains(lastCmd, "--build-arg ARG1=value1") {
			t.Errorf("Build arg not found in command: %s", lastCmd)
		}
	})
}

func TestBuilder_CommandFormat(t *testing.T) {
	mock := NewMockSSHClient()
	builder := NewBuilder(mock)

	t.Run("Verify Command Format", func(t *testing.T) {
		mock.SetOutput("podman build", "")
		opts := BuildOptions{
			File:        "Dockerfile",
			ContextPath: "./app",
			Tag:        "test:latest",
			Target:     "production",
			NoCache:    true,
		}
		
		if err := builder.Build(opts); err != nil {
			t.Errorf("Build failed: %v", err)
		}

		commands := mock.GetCommands()
		expectedCmd := "podman build -f Dockerfile -t test:latest --target production --no-cache ./app"
		if !strings.Contains(commands[len(commands)-1], expectedCmd) {
			t.Errorf("Command format mismatch\nExpected: %s\nGot: %s", expectedCmd, commands[len(commands)-1])
		}
	})
}

func TestBuilder_SSHConnection(t *testing.T) {
	mock := NewMockSSHClient()
	builder := NewBuilder(mock)

	t.Run("SSH Client Used Properly", func(t *testing.T) {
		mock.SetOutput("podman build", "")
		opts := BuildOptions{
			File:        "Dockerfile",
			ContextPath: ".",
			Tag:        "test:latest",
		}

		// Verify the builder uses the SSH client we provided
		if builder.ssh != mock {
			t.Error("Builder not using provided SSH client")
		}
		
		if err := builder.Build(opts); err != nil {
			t.Errorf("Build failed: %v", err)
		}

		commands := mock.GetCommands()
		if len(commands) == 0 {
			t.Error("No commands were executed through SSH client")
		}
	})
}

func TestBuilder_BuildWithAllFeatures(t *testing.T) {
	mock := NewMockSSHClient()
	builder := NewBuilder(mock)

	t.Run("Build With All Features Combined", func(t *testing.T) {
		mock.SetOutput("podman build", "")
		opts := BuildOptions{
			File:        "custom.Dockerfile",
			ContextPath: "./app",
			Tag:        "test:latest",
			BuildArgs: map[string]string{
				"ARG1": "value1",
			},
			Environment: map[string]string{
				"ENV1": "env1",
			},
			Volumes:  []string{"/data:/data"},
			Target:   "prod",
			NoCache:  true,
			Pull:     "always",
			Platform: []string{"linux/amd64"},
		}

		if err := builder.Build(opts); err != nil {
			t.Errorf("Build failed: %v", err)
		}

		commands := mock.GetCommands()
		expectedParts := []string{
			"podman build",
			"-f custom.Dockerfile",
			"-t test:latest",
			"--build-arg ARG1=value1",
			"--env ENV1=env1",
			"--volume /data:/data",
			"--target prod",
			"--no-cache",
			"--pull always",
			"--platform linux/amd64",
			"./app",
		}

		for _, part := range expectedParts {
			if !strings.Contains(commands[len(commands)-1], part) {
				t.Errorf("Expected command to contain %q, got: %s", part, commands[len(commands)-1])
			}
		}
	})
}

func TestBuilder_EmptyBuildOptions(t *testing.T) {
	mock := NewMockSSHClient()
	builder := NewBuilder(mock)

	t.Run("Build With Minimal Options", func(t *testing.T) {
		mock.SetOutput("podman build", "")
		opts := BuildOptions{
			ContextPath: ".",
			Tag:        "test:latest",
		}

		if err := builder.Build(opts); err != nil {
			t.Errorf("Build failed: %v", err)
		}

		commands := mock.GetCommands()
		lastCmd := commands[len(commands)-1]
		expectedCmd := "podman build -t test:latest ."
		if lastCmd != expectedCmd {
			t.Errorf("Command format mismatch\nExpected: %s\nGot: %s", expectedCmd, lastCmd)
		}
	})
}

func TestBuilder_ValidateOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    BuildOptions
		wantErr bool
		errMsg  string
	}{
		{
			name: "invalid platform format",
			opts: BuildOptions{
				Platform: []string{"invalid"},
			},
			wantErr: true,
			errMsg:  "invalid platform format",
		},
		{
			name: "invalid pull value",
			opts: BuildOptions{
				Pull: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid pull value",
		},
		// Add more validation test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilder(NewMockSSHClient())
			err := builder.validateOptions(tt.opts)
			
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %v", tt.errMsg, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestBuilder_EdgeCases(t *testing.T) {
	t.Run("Build with very long command", func(t *testing.T) {
		ssh := NewMockSSHClient()
		builder := NewBuilder(ssh)

		// Set up mock responses
		ssh.SetOutput("test -d .", "")
		ssh.SetOutput("test -f ./Dockerfile", "")
		ssh.SetOutput("podman build", "")

		buildArgs := make(map[string]string)
		for i := 0; i < 100; i++ {
			buildArgs[fmt.Sprintf("ARG%d", i)] = "value"
		}
		
		opts := BuildOptions{
			File:        "Dockerfile",
			ContextPath: ".",
			Tag:        "test:latest",
			BuildArgs:   buildArgs,
		}
		
		err := builder.Build(opts)
		assert.NoError(t, err)

		commands := ssh.GetCommands()
		var buildCmd string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, "podman build") {
				buildCmd = cmd
				break
			}
		}
		assert.NotEmpty(t, buildCmd)
	})

	t.Run("Build with special characters", func(t *testing.T) {
		ssh := NewMockSSHClient()
		builder := NewBuilder(ssh)

		// Set up mock responses
		ssh.SetOutput("test -d ./path with spaces", "")
		ssh.SetOutput("test -f ./path with spaces/Docker file with spaces.dockerfile", "")
		ssh.SetOutput("podman build", "")

		opts := BuildOptions{
			File:        "Docker file with spaces.dockerfile",
			ContextPath: "./path with spaces",
			Tag:        "test:latest",
			BuildArgs: map[string]string{
				"ARG": "value with spaces and $pecial chars",
			},
		}
		
		err := builder.Build(opts)
		assert.NoError(t, err)

		commands := ssh.GetCommands()
		var buildCmd string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, "podman build") {
				buildCmd = cmd
				break
			}
		}
		assert.NotEmpty(t, buildCmd)
		assert.Contains(t, buildCmd, "Docker file with spaces.dockerfile")
	})
}

func TestBuilder_Concurrent(t *testing.T) {
	mock := NewMockSSHClient()
	builder := NewBuilder(mock)

	t.Run("Concurrent builds", func(t *testing.T) {
		var wg sync.WaitGroup
		errChan := make(chan error, 10)
		
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				opts := BuildOptions{
					File:        "Dockerfile",
					ContextPath: ".",
					Tag: fmt.Sprintf("test%d:latest", i),
				}
				if err := builder.Build(opts); err != nil {
					errChan <- err
				}
			}(i)
		}
		
		wg.Wait()
		close(errChan)
		
		for err := range errChan {
			t.Errorf("concurrent build error: %v", err)
		}
	})
}

func TestBuilder_ResourceCleanup(t *testing.T) {
	mock := NewMockSSHClient()
	builder := NewBuilder(mock)

	t.Run("Cleanup after failed build", func(t *testing.T) {
		// Set up mock responses
		mock.SetOutput("test -d .", "")
		mock.SetOutput("test -f ./Dockerfile", "")
		mock.SetError("podman build", fmt.Errorf("build failed"))
		// Set up success response for cleanup command
		mock.SetOutput("podman rm", "")
		
		opts := BuildOptions{
			File:        "Dockerfile",
			ContextPath: ".",
			Tag:        "test:latest", // Add required tag
			ForceRm:    true,
		}
		
		err := builder.Build(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "build failed")
		
		// Verify cleanup commands were executed
		commands := mock.GetCommands()
		foundCleanup := false
		for _, cmd := range commands {
			if strings.Contains(cmd, "podman rm $(podman ps -a -q -f status=exited)") {
				foundCleanup = true
				break
			}
		}
		assert.True(t, foundCleanup, "cleanup commands not found after failed build")
	})
}
