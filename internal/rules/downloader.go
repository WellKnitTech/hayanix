package rules

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type GitHubDownloader struct {
	client *http.Client
}

func NewGitHubDownloader() *GitHubDownloader {
	return &GitHubDownloader{
		client: &http.Client{},
	}
}

func (gd *GitHubDownloader) DownloadRepositoryRules(repoURL, branch, targetDir string, rulePaths []string) error {
	// Convert GitHub URL to archive URL
	archiveURL := gd.getArchiveURL(repoURL, branch)
	
	// Download the archive
	archiveData, err := gd.downloadArchive(archiveURL)
	if err != nil {
		return fmt.Errorf("failed to download archive: %w", err)
	}

	// Extract relevant files
	return gd.extractRules(archiveData, targetDir, rulePaths)
}

func (gd *GitHubDownloader) getArchiveURL(repoURL, branch string) string {
	// Convert https://github.com/user/repo to https://github.com/user/repo/archive/branch.tar.gz
	if strings.HasSuffix(repoURL, ".git") {
		repoURL = strings.TrimSuffix(repoURL, ".git")
	}
	
	return fmt.Sprintf("%s/archive/%s.tar.gz", repoURL, branch)
}

func (gd *GitHubDownloader) downloadArchive(url string) ([]byte, error) {
	resp, err := gd.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download archive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download archive: status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (gd *GitHubDownloader) extractRules(archiveData []byte, targetDir string, rulePaths []string) error {
	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Create a map of paths we want to extract
	wantedPaths := make(map[string]bool)
	for _, path := range rulePaths {
		wantedPaths[path] = true
	}

	// Extract tar.gz archive
	gzReader, err := gzip.NewReader(bytes.NewReader(archiveData))
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Check if this file matches our wanted paths
		if gd.shouldExtractFile(header.Name, wantedPaths) {
			if err := gd.extractFile(tarReader, header, targetDir); err != nil {
				return fmt.Errorf("failed to extract file %s: %w", header.Name, err)
			}
		}
	}

	return nil
}

func (gd *GitHubDownloader) shouldExtractFile(filePath string, wantedPaths map[string]bool) bool {
	// Remove the repository name prefix (e.g., "repo-name-abc123/")
	parts := strings.Split(filePath, "/")
	if len(parts) < 2 {
		return false
	}
	
	// Reconstruct the path without the repo prefix
	relativePath := strings.Join(parts[1:], "/")
	
	// Check if this path starts with any of our wanted paths
	for wantedPath := range wantedPaths {
		if strings.HasPrefix(relativePath, wantedPath+"/") || relativePath == wantedPath {
			// Only extract YAML files
			return strings.HasSuffix(filePath, ".yml") || strings.HasSuffix(filePath, ".yaml")
		}
	}
	
	return false
}

func (gd *GitHubDownloader) extractFile(reader io.Reader, header *tar.Header, targetDir string) error {
	// Create the target file path
	targetPath := filepath.Join(targetDir, header.Name)
	
	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	file, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy the file content
	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// Alternative method using GitHub API for individual file downloads
type GitHubAPIDownloader struct {
	client *http.Client
	token  string
}

func NewGitHubAPIDownloader(token string) *GitHubAPIDownloader {
	return &GitHubAPIDownloader{
		client: &http.Client{},
		token:  token,
	}
}

func (gad *GitHubAPIDownloader) DownloadFile(repoOwner, repoName, filePath, targetPath string) error {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/master/%s", repoOwner, repoName, filePath)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if gad.token != "" {
		req.Header.Set("Authorization", "token "+gad.token)
	}

	resp, err := gad.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: status %d", resp.StatusCode)
	}

	// Create target directory
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	file, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy the content
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}
