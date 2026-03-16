package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/bensch98/arcane/internal/ui"
	"github.com/spf13/cobra"
)

const (
	repoOwner = "bensch98"
	repoName  = "arcane"
)

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade the arcane CLI to the latest release",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip registry loading — upgrade doesn't need it.
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Current version: %s\n", ui.Cyan(Version))

		// Fetch latest release from GitHub.
		rel, err := fetchLatestRelease()
		if err != nil {
			ui.Die("Failed to check for updates: %v", err)
		}

		latest := strings.TrimPrefix(rel.TagName, "v")
		current := strings.TrimPrefix(Version, "v")

		fmt.Printf("Latest version:  %s\n", ui.Cyan(latest))

		if current == latest {
			fmt.Println(ui.Green("Already up to date."))
			return
		}

		// Find the right binary for this platform.
		wantName := fmt.Sprintf("arcane-%s-%s", runtime.GOOS, runtime.GOARCH)
		var downloadURL string
		for _, a := range rel.Assets {
			if a.Name == wantName {
				downloadURL = a.BrowserDownloadURL
				break
			}
		}
		if downloadURL == "" {
			ui.Die("No binary found for %s/%s in release %s", runtime.GOOS, runtime.GOARCH, rel.TagName)
		}

		fmt.Printf("\nUpgrading %s → %s ...\n", ui.Yellow(current), ui.Green(latest))

		// Download the new binary to a temp file.
		tmpFile, err := downloadBinary(downloadURL)
		if err != nil {
			ui.Die("Download failed: %v", err)
		}
		defer os.Remove(tmpFile)

		// Replace the current binary.
		execPath, err := os.Executable()
		if err != nil {
			ui.Die("Cannot determine executable path: %v", err)
		}

		if err := replaceBinary(tmpFile, execPath); err != nil {
			ui.Die("Failed to replace binary: %v", err)
		}

		fmt.Println(ui.Green("Upgrade complete."))
	},
}

func fetchLatestRelease() (*ghRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func downloadBinary(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned %s", resp.Status)
	}

	tmp, err := os.CreateTemp("", "arcane-upgrade-*")
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}

	return tmp.Name(), nil
}

func replaceBinary(src, dst string) error {
	// Preserve permissions of the existing binary.
	info, err := os.Stat(dst)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Write to a temp file next to the target, then atomic rename.
	tmpDst := dst + ".new"
	out, err := os.OpenFile(tmpDst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, srcFile); err != nil {
		out.Close()
		os.Remove(tmpDst)
		return err
	}
	out.Close()

	return os.Rename(tmpDst, dst)
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
