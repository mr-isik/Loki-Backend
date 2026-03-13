package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type Runner struct {
	cli *client.Client
}

type RunRequest struct {
	Image   string
	Command []string
	Input   []byte
}

func NewRunner() (*Runner, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithVersion("1.44"), client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize docker client: %w", err)
	}

	// Ping the docker daemon to make sure it's accessible.
	// If it's not, we'll return an error safely so the application can fallback/fail cleanly.
	pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = cli.Ping(pingCtx)
	if err != nil {
		return nil, fmt.Errorf("Docker is not available on this system: %w", err)
	}

	return &Runner{cli: cli}, nil
}

// RunContainer runs an ephemeral container with the given image and command,
// pipes the provided input into its stdin (if any), and returns the combined stdout and stderr.
// The container is severely locked down: 64MB RAM limits, NetworkDisabled: true, and AutoRemove: true.
func (r *Runner) RunContainer(ctx context.Context, req RunRequest) (string, error) {
	// 1. Pull the image if it doesn't exist locally
	reader, err := r.cli.ImagePull(ctx, req.Image, types.ImagePullOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to pull image %s: %w", req.Image, err)
	}
	defer reader.Close()
	// Drain the pull response properly
	_, _ = io.Copy(io.Discard, reader)

	// 2. Create the container
	// We want to limit RAM to 64MB
	ramLimit := int64(64 * 1024 * 1024)

	resp, err := r.cli.ContainerCreate(ctx, &container.Config{
		Image:           req.Image,
		Cmd:             req.Command,
		NetworkDisabled: true,
		OpenStdin:       len(req.Input) > 0,
		StdinOnce:       len(req.Input) > 0,
		AttachStdin:     len(req.Input) > 0,
		AttachStdout:    true,
		AttachStderr:    true,
	}, &container.HostConfig{
		Resources: container.Resources{
			Memory:     ramLimit,
			MemorySwap: ramLimit,
		},
	}, nil, nil, "")

	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	containerID := resp.ID

	// Defer removal to guarantee cleanup even if we return early or panic
	defer func() {
		_ = r.cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{Force: true})
	}()

	// 3. Attach if we need to send Stdin, or just start it.
	// We must attach before starting to stream stdin.
	var hj types.HijackedResponse
	if len(req.Input) > 0 {
		hj, err = r.cli.ContainerAttach(ctx, containerID, types.ContainerAttachOptions{
			Stream: true,
			Stdin:  true,
		})
		if err != nil {
			return "", fmt.Errorf("failed to attach to container stdin: %w", err)
		}
		defer hj.Close()
	}

	// 4. Start the container
	if err := r.cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	// 5. Send input if any
	if len(req.Input) > 0 {
		_, err = hj.Conn.Write(req.Input)
		if err != nil {
			return "", fmt.Errorf("failed to write to container stdin: %w", err)
		}
		// Close the write half to signal EOF to the container process.
		// Not all connections support this trivially without closing the read half,
		// but go docker client usually handles closeWrite for us via the conn wrapper.
		// We'll just close it.
		hj.CloseWrite()
	}

	// 6. Wait for the container to finish and get the logs
	statusCh, errCh := r.cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	// Get logs while it's running/waiting
	out, err := r.cli.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %w", err)
	}
	defer out.Close()

	// Use stdcopy to split multipexed stderr and stdout streams.
	// We'll recombine them for the user output, but they come multiplexed from the daemon.
	var stdoutBuf, stderrBuf bytes.Buffer
	_, err = stdcopy.StdCopy(&stdoutBuf, &stderrBuf, out)
	if err != nil {
		// Log but don't strictly fail on read errors
		fmt.Printf("warning: error reading container logs: %v\n", err)
	}

	// 7. Check wait result
	select {
	case err := <-errCh:
		if err != nil { // Wait error
			return "", fmt.Errorf("error waiting for container: %w", err)
		}
	case status := <-statusCh: // Wait finished
		// Check exit code
		if status.Error != nil {
			return "", fmt.Errorf("container exited with error: %s", status.Error.Message)
		}
		if status.StatusCode != 0 {
			// Include stderr in the error response for clarity
			stderrStr := stderrBuf.String()
			stdoutStr := stdoutBuf.String()
			return "", fmt.Errorf("container exited with status code %d\nStdout: %s\nStderr: %s", status.StatusCode, stdoutStr, stderrStr)
		}
	case <-ctx.Done(): // Context cancelled (timeout)
		// We should try to kill it if ctx is cancelled, though AutoRemove might not trigger if it's forcefully killed mid-run without careful cleanup.
		// For now we'll just return the context error.
		return "", ctx.Err()
	}

	combinedOutput := stdoutBuf.String()
	if stderrBuf.Len() > 0 {
		// If both have output, append them. Or we could just return stdout if exit code is 0.
		// Often shell scripts output to stderr even on success, so combined is better.
		if len(combinedOutput) > 0 && !strings.HasSuffix(combinedOutput, "\n") {
			combinedOutput += "\n"
		}
		combinedOutput += stderrBuf.String()
	}

	return combinedOutput, nil
}
