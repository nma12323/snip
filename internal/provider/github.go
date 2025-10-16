package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// GitHub structures
type GitTreeResponse struct {
	SHA       string        `json:"sha"`
	Tree      []GitTreeItem `json:"tree"`
	Truncated bool          `json:"truncated"`
}

type GitTreeItem struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"` // "blob" or "tree"
	SHA  string `json:"sha"`
	Size int    `json:"size,omitempty"`
	URL  string `json:"url"`
}

// Content API response for a single file
type contentResponse struct {
	Type     string `json:"type"`
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
	SHA      string `json:"sha"`
}

type GitHubClient struct {
	ghToken   string
	userAgent string
	client    *http.Client
}

func NewGitHubClient() *GitHubClient {
	return &GitHubClient{
		ghToken:   os.Getenv("GITHUB_TOKEN"),
		userAgent: "Snip/1.0",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (g *GitHubClient) Token() string {
	return g.ghToken
}

// IsPrivate checks repo visibility. Retries with token if necessary.
func (g *GitHubClient) IsPrivate(owner, repo string) (bool, error) {
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	resp, err := g.apiRequest(api)
	if err != nil {
		return false, fmt.Errorf("cannot determine visibility; likely private or nonexistent\n%v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("github API error: %s", strings.TrimSpace(string(body)))
	}

	var data struct {
		Private bool `json:"private"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return false, err
	}
	return data.Private, nil
}

func (g *GitHubClient) GetDefaultBranch(owner, repo string) (string, error) {
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	resp, err := g.apiRequest(api)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("github API error: %s", strings.TrimSpace(string(body)))
	}

	var info struct {
		DefaultBranch string `json:"default_branch"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}
	if info.DefaultBranch == "" {
		return "", errors.New("default branch not found")
	}
	return info.DefaultBranch, nil
}

func (g *GitHubClient) ListTree(owner, repo, branch string) ([]GitTreeItem, error) {
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1",
		owner, repo, url.PathEscape(branch))

	resp, err := g.apiRequest(api)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s", strings.TrimSpace(string(body)))
	}

	var tr struct {
		Tree      []GitTreeItem `json:"tree"`
		Truncated bool          `json:"truncated"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}

	if tr.Truncated {
		fmt.Println("⚠️ GitHub Tree API response truncated.")
		fmt.Println("👉 The repository is large, and GitHub limited the response size.")
		fmt.Println("💡 You can retry using a GitHub token")
	}

	return tr.Tree, nil
}

func (g *GitHubClient) GetFileContent(owner, repo, branch, path string) (string, error) {
	api := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", owner, repo, url.PathEscape(path), url.QueryEscape(branch))
	resp, err := g.apiRequest(api)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("github API error: %s", strings.TrimSpace(string(body)))
	}

	var cr contentResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return "", err
	}
	if cr.Encoding != "base64" {
		return "", fmt.Errorf("unexpected encoding: %s", cr.Encoding)
	}
	return cr.Content, nil
}

func (g *GitHubClient) apiRequest(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", g.userAgent)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	if g.ghToken != "" {
		req.Header.Set("Authorization", "token "+g.ghToken)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 403 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		bodyStr := strings.ToLower(string(body))

		// Abuse detection or rate limit
		if strings.Contains(bodyStr, "abuse") {
			return nil, fmt.Errorf("⚠️  GitHub API abuse detection triggered. Please slow down and try again later")
		}

		reset := resp.Header.Get("X-RateLimit-Reset")
		var waitMsg string
		if reset != "" {
			if ts, err := strconv.ParseInt(reset, 10, 64); err == nil {
				resetTime := time.Unix(ts, 0)
				waitMsg = fmt.Sprintf("Try again after %s", resetTime.Format(time.RFC1123))
			}
		}

		if g.ghToken == "" {
			return nil, fmt.Errorf("⚠️  Unauthenticated GitHub API rate limit exceeded.\n%s\nTip: set GITHUB_TOKEN to increase your rate limit", waitMsg)
		}
		return nil, fmt.Errorf("⚠️  Authenticated GitHub API rate limit exceeded.\n%s", waitMsg)
	}

	// Handle other HTTP errors
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("GitHub API error (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return resp, nil
}

func SimpleGet(rawURL string) (*http.Response, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("GET", rawURL, nil)
	req.Header.Set("User-Agent", "Snip/1.0")
	return client.Do(req)
}
