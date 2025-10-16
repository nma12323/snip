package cmd

import (
	"strings"

	"github.com/mukailasam/snip/internal/provider"
)

type treeBlob struct {
	Path string
	Size int
}

func findDirsByName(tree []provider.GitTreeItem, name string) []string {
	var matches []string
	for _, it := range tree {
		if it.Type == "tree" {
			// check last segment
			parts := strings.Split(it.Path, "/")
			if parts[len(parts)-1] == name {
				matches = append(matches, it.Path)
			}
		}
	}
	return matches
}

func filterBlobsUnder(tree []provider.GitTreeItem, dir string) []treeBlob {
	prefix := dir + "/"
	var out []treeBlob
	for _, it := range tree {
		if it.Type == "blob" && strings.HasPrefix(it.Path, prefix) {
			out = append(out, treeBlob{Path: it.Path, Size: it.Size})
		}
	}
	return out
}

func findFilesByName(tree []provider.GitTreeItem, name string) []provider.GitTreeItem {
	var out []provider.GitTreeItem
	for _, it := range tree {
		if it.Type == "blob" {
			parts := strings.Split(it.Path, "/")
			if parts[len(parts)-1] == name {
				out = append(out, it)
			}
		}
	}
	return out
}
