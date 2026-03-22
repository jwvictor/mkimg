package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const releasesURL = "https://api.github.com/repos/jwvictor/mkimg/releases/latest"

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("mkimg %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for updates and self-update",
	Long: `Check GitHub releases for a newer version of mkimg.
If a newer version is found, download and replace the current binary.

Use --check to only check without installing.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		checkOnly, _ := cmd.Flags().GetBool("check")

		fmt.Printf("Current version: %s\n", version)

		// Fetch latest release
		rel, err := fetchLatestRelease()
		if err != nil {
			return fmt.Errorf("failed to check for updates: %w", err)
		}

		latest := strings.TrimPrefix(rel.TagName, "v")
		fmt.Printf("Latest version:  %s\n", latest)

		if latest == version || version == "dev" && !checkOnly {
			if version == "dev" {
				fmt.Println("\nRunning a dev build. Use --check to see the latest release.")
				return nil
			}
			fmt.Println("\nAlready up to date.")
			return nil
		}

		if checkOnly {
			if latest != version {
				fmt.Println("\nUpdate available. Run 'mkimg update' to install.")
			}
			return nil
		}

		// Find the right asset for this platform
		assetName := fmt.Sprintf("mkimg_%s_%s_%s.tar.gz", latest, runtime.GOOS, runtime.GOARCH)
		var downloadURL string
		for _, a := range rel.Assets {
			if a.Name == assetName {
				downloadURL = a.BrowserDownloadURL
				break
			}
		}
		if downloadURL == "" {
			return fmt.Errorf("no binary found for %s/%s in release %s", runtime.GOOS, runtime.GOARCH, rel.TagName)
		}

		// Download
		fmt.Printf("\nDownloading %s...\n", assetName)
		binary, err := downloadAndExtract(downloadURL)
		if err != nil {
			return fmt.Errorf("download failed: %w", err)
		}

		// Replace current binary
		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("cannot find current binary path: %w", err)
		}

		fmt.Printf("Replacing %s...\n", execPath)
		if err := replaceBinary(execPath, binary); err != nil {
			return fmt.Errorf("update failed: %w", err)
		}

		fmt.Printf("Updated to mkimg %s\n", latest)
		return nil
	},
}

func fetchLatestRelease() (*ghRelease, error) {
	resp, err := http.Get(releasesURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func downloadAndExtract(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("download returned %d", resp.StatusCode)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("tar: %w", err)
		}
		if hdr.Name == "mkimg" {
			return io.ReadAll(tr)
		}
	}

	return nil, fmt.Errorf("mkimg binary not found in archive")
}

func replaceBinary(path string, newBinary []byte) error {
	// Write to a temp file next to the target, then rename (atomic on same fs)
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, newBinary, 0755); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
}

func init() {
	updateCmd.Flags().Bool("check", false, "Only check for updates, don't install")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
}
