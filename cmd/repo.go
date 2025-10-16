package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mukailasam/snip/internal/provider"
	"github.com/mukailasam/snip/utils"
	"github.com/spf13/cobra"
)

var (
	flagDir    string
	flagFile   string
	flagBranch string
	flagDest   string
)

func init() {
	repoCmd.Flags().StringVar(&flagDir, "dir", "", "Name of directory to snip (searches entire repo)")
	repoCmd.Flags().StringVar(&flagFile, "file", "", "Name of file to snip (searches entire repo)")
	repoCmd.Flags().StringVar(&flagBranch, "branch", "", "Branch to use (optional, default = repo default branch)")
	repoCmd.Flags().StringVar(&flagDest, "dest", ".", "Destination directory to write files")
	rootCmd.AddCommand(repoCmd)
}

var repoCmd = &cobra.Command{
	Use:   "repo <repo-url>",
	Short: "Snip a directory or file from a repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoURL := args[0]

		if flagDir == "" && flagFile == "" {
			return fmt.Errorf("please provide --dir or --file")
		}
		if flagDir != "" && flagFile != "" {
			return fmt.Errorf("provide only one of --dir or --file")
		}

		owner, repo, host, err := utils.ParseRepoURL(repoURL)
		if err != nil {
			return err
		}
		if !strings.Contains(host, "github.com") {
			return fmt.Errorf("only github.com is supported")
		}

		gh := provider.NewGitHubClient()

		isPrivate, err := gh.IsPrivate(owner, repo)
		if err != nil {
			return fmt.Errorf("⚠️  %v:\nDetails: Ensure the repository URL is correct and accessible", err)
		}

		if isPrivate {
			fmt.Println("🔒 Private repository detected.")
			if gh.Token() == "" {
				return fmt.Errorf("private repo access requires GITHUB_TOKEN environment variable")
			}
		} else {
			fmt.Println("🌍 Public repository detected.")
		}

		// determine branch
		branch := flagBranch
		if branch == "" {
			branch, err = gh.GetDefaultBranch(owner, repo)
			if err != nil {
				fmt.Printf("⚠️  Could not detect default branch, using 'master': %v\n", err)
				branch = "master"
			}
		}

		fmt.Printf("🔍 Listing repository tree for %s/%s (branch: %s)...\n", owner, repo, branch)
		tree, err := gh.ListTree(owner, repo, branch)
		if err != nil {
			return fmt.Errorf("failed to list repo tree: %w", err)
		}

		if flagDir != "" {
			// find directories whose last path element equals flagDir
			matches := findDirsByName(tree, flagDir)
			if len(matches) == 0 {
				return fmt.Errorf("no directory named %q found in repo", flagDir)
			}
			if len(matches) > 1 {
				fmt.Printf("⚠️  Multiple directories named %q found, selecting the first match: %s\n", flagDir, matches[0])
			}
			target := matches[0]
			fmt.Printf("📦 Snipping directory: %s\n", target)
			// collect blobs under target prefix
			blobs := filterBlobsUnder(tree, target)
			if len(blobs) == 0 {
				fmt.Println("⚠️ Directory is empty.")
				return nil
			}
			for _, b := range blobs {
				relPath := b.Path[len(target)+1:] // path relative to target
				localPath := filepath.Join(flagDest, filepath.Base(target), relPath)
				if isPrivate {
					// use GitHub contents API to get file content (base64)
					content, err := gh.GetFileContent(owner, repo, branch, b.Path)
					if err != nil {
						return err
					}
					if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
						return err
					}
					data, _ := base64.StdEncoding.DecodeString(content)
					if err := os.WriteFile(localPath, data, 0o644); err != nil {
						return err
					}
					fmt.Println("⬇️  Downloaded (private):", localPath)
				} else {
					raw := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, branch, b.Path)
					if err := utils.DownloadToFile(raw, localPath); err != nil {
						return err
					}
					fmt.Println("⬇️  Downloaded:", localPath)
				}
			}
			fmt.Println("✅ Done.")
			return nil
		}

		// file mode
		if flagFile != "" {
			matches := findFilesByName(tree, flagFile)
			if len(matches) == 0 {
				return fmt.Errorf("no file named %q found in repo", flagFile)
			}
			if len(matches) > 1 {
				fmt.Printf("⚠️  Multiple files named %q found, selecting the first match: %s\n", flagFile, matches[0].Path)
			}
			target := matches[0]
			localPath := filepath.Join(flagDest, filepath.Base(target.Path))
			if isPrivate {
				content, err := gh.GetFileContent(owner, repo, branch, target.Path)
				if err != nil {
					return err
				}
				data, _ := base64.StdEncoding.DecodeString(content)
				if err := os.WriteFile(localPath, data, 0o644); err != nil {
					return err
				}
				fmt.Println("⬇️  Downloaded (private):", localPath)
			} else {
				raw := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, branch, target.Path)
				if err := utils.DownloadToFile(raw, localPath); err != nil {
					return err
				}
				fmt.Println("⬇️  Downloaded:", localPath)
			}
			fmt.Println("✅ Done.")
			return nil
		}

		return nil
	},
}
