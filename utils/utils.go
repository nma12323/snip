package utils

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/mukailasam/snip/internal/provider"
)

// ParseRepoURL extracts owner, repo and host. Accepts URLs like:
//   - https://github.com/owner/repo
//   - github.com/owner/repo
func ParseRepoURL(raw string) (owner, repo, host string, err error) {
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid repo URL: %w", err)
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("invalid repo URL, expected host/owner/repo")
	}
	return parts[0], parts[1], u.Hostname(), nil
}

func DownloadToFile(url, localPath string) error {
	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		return err
	}
	out, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := provider.SimpleGet(url) // provider helper below
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to download: %s", string(body))
	}
	_, err = io.Copy(out, resp.Body)
	return err
}
