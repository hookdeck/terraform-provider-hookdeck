package project

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"syscall"
	"time"

	"terraform-provider-hookdeck/internal/sdkclient"
)

const (
	codeNotProvisioned = "OUTPOST_NOT_PROVISIONED"
	codeConfigLocked   = "OUTPOST_CONFIG_LOCKED"

	lockRetries      = 3
	transientRetries = 3
)

var (
	pollInterval    = 2 * time.Second
	pollMaxInterval = 15 * time.Second
	lockRetryDelay  = 5 * time.Second
)

const (
	resourceKeyHint   = "Check that the API key is an organization API key with access to the project."
	dataSourceKeyHint = "Use an organization API key with access to the project, or the project's own API key."
)

func outpostAPI(client sdkclient.Client, projectID string) sdkclient.Client {
	return client.Outpost().WithProject(projectID, resourceKeyHint)
}

func getConfig(ctx context.Context, api sdkclient.Client) (map[string]*string, error) {
	var config map[string]*string
	if err := api.Do(ctx, http.MethodGet, "/config", nil, nil, &config); err != nil {
		return nil, err
	}
	return config, nil
}

func patchConfig(ctx context.Context, api sdkclient.Client, patch map[string]*string) (map[string]*string, error) {
	locked, transient := 0, 0
	for {
		var config map[string]*string
		err := api.Do(ctx, http.MethodPatch, "/config", patch, nil, &config)
		if err == nil {
			return config, nil
		}
		// The patch holds absolute values, so sending it again is safe.
		switch {
		case sdkclient.IsCode(err, http.StatusConflict, codeConfigLocked) && locked < lockRetries:
			locked++
		case retryableWait(err) && !sdkclient.IsCode(err, http.StatusConflict, codeNotProvisioned) && transient < transientRetries:
			transient++
		default:
			return nil, err
		}
		if err := sleep(ctx, lockRetryDelay); err != nil {
			return nil, err
		}
	}
}

func waitProvisioned(ctx context.Context, api sdkclient.Client, projectID string, timeout time.Duration) error {
	delay := pollInterval
	for {
		_, err := getConfig(ctx, api)
		if err == nil {
			return nil
		}
		if !retryableWait(err) {
			return err
		}
		if sleep(ctx, delay) != nil {
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return provisionTimeoutError{projectID: projectID, timeout: timeout}
			}
			return ctx.Err()
		}
		delay = min(delay*2, pollMaxInterval)
	}
}

// retryableWait retries what clears while a project provisions: 5xx, 429, not
// provisioned, and transient network errors. A host that doesn't resolve or a
// bad certificate is a configuration error and fails at once.
func retryableWait(err error) bool {
	var apiErr *sdkclient.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Status >= http.StatusInternalServerError ||
			apiErr.Status == http.StatusTooManyRequests ||
			sdkclient.IsCode(err, http.StatusConflict, codeNotProvisioned)
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF)
}

type provisionTimeoutError struct {
	projectID string
	timeout   time.Duration
}

func (e provisionTimeoutError) Error() string {
	return fmt.Sprintf("Outpost project %s was created but is not ready after %s. It is kept in state and will be replaced on the next apply; raise timeouts.create if provisioning is slow.", e.projectID, e.timeout)
}

func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
