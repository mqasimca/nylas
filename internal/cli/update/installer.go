package update

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const binaryName = "nylas"

// getAssetName returns the expected asset name for the current platform.
func getAssetName(version string) string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}

	return fmt.Sprintf("%s_%s_%s_%s%s", binaryName, version, goos, goarch, ext)
}

// downloadFile downloads a file from URL to a temporary file.
func downloadFile(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", "nylas-cli")

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "nylas-update-*")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		return "", fmt.Errorf("write temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", fmt.Errorf("close temp file: %w", err)
	}

	return tmpFile.Name(), nil
}

// downloadChecksums downloads and parses the checksums.txt file.
func downloadChecksums(ctx context.Context, url string) (map[string]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", "nylas-cli")

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download checksums: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("checksums download failed with status %d", resp.StatusCode)
	}

	checksums := make(map[string]string)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Format: "checksum  filename" (two spaces)
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			checksums[parts[1]] = parts[0]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("parse checksums: %w", err)
	}

	return checksums, nil
}

// verifyChecksum verifies the SHA256 checksum of a file.
func verifyChecksum(filePath, expected string) (bool, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, fmt.Errorf("hash file: %w", err)
	}

	actual := hex.EncodeToString(h.Sum(nil))
	return strings.EqualFold(actual, expected), nil
}

// extractBinary extracts the binary from the archive.
func extractBinary(archivePath, goos string) (string, error) {
	if goos == "windows" {
		return extractFromZip(archivePath)
	}
	return extractFromTarGz(archivePath)
}

// extractFromTarGz extracts the binary from a tar.gz archive.
func extractFromTarGz(archivePath string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("open archive: %w", err)
	}
	defer func() { _ = f.Close() }()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("create gzip reader: %w", err)
	}
	defer func() { _ = gzr.Close() }()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read tar: %w", err)
		}

		// Look for the binary (could be "nylas" or "nylas.exe")
		baseName := filepath.Base(header.Name)
		if baseName == binaryName || baseName == binaryName+".exe" {
			tmpFile, err := os.CreateTemp("", "nylas-binary-*")
			if err != nil {
				return "", fmt.Errorf("create temp file: %w", err)
			}

			if _, err := io.Copy(tmpFile, tr); err != nil {
				_ = tmpFile.Close()
				_ = os.Remove(tmpFile.Name())
				return "", fmt.Errorf("extract binary: %w", err)
			}

			if err := tmpFile.Close(); err != nil {
				_ = os.Remove(tmpFile.Name())
				return "", fmt.Errorf("close temp file: %w", err)
			}

			return tmpFile.Name(), nil
		}
	}

	return "", fmt.Errorf("binary not found in archive")
}

// extractFromZip extracts the binary from a zip archive.
func extractFromZip(archivePath string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("open zip: %w", err)
	}
	defer func() { _ = r.Close() }()

	for _, f := range r.File {
		baseName := filepath.Base(f.Name)
		if baseName == binaryName || baseName == binaryName+".exe" {
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("open file in zip: %w", err)
			}

			tmpFile, err := os.CreateTemp("", "nylas-binary-*")
			if err != nil {
				_ = rc.Close()
				return "", fmt.Errorf("create temp file: %w", err)
			}

			if _, err := io.Copy(tmpFile, rc); err != nil {
				_ = tmpFile.Close()
				_ = rc.Close()
				_ = os.Remove(tmpFile.Name())
				return "", fmt.Errorf("extract binary: %w", err)
			}

			_ = rc.Close()

			if err := tmpFile.Close(); err != nil {
				_ = os.Remove(tmpFile.Name())
				return "", fmt.Errorf("close temp file: %w", err)
			}

			return tmpFile.Name(), nil
		}
	}

	return "", fmt.Errorf("binary not found in archive")
}

// getCurrentBinaryPath returns the path to the current binary.
func getCurrentBinaryPath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("get executable path: %w", err)
	}

	// Resolve symlinks to get the real path
	realPath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		return "", fmt.Errorf("resolve symlinks: %w", err)
	}

	return realPath, nil
}

// isHomebrewInstall checks if the binary was installed via Homebrew.
func isHomebrewInstall() bool {
	exePath, err := getCurrentBinaryPath()
	if err != nil {
		return false
	}

	// Homebrew installs to /opt/homebrew/Cellar or /usr/local/Cellar
	return strings.Contains(exePath, "/Cellar/") ||
		strings.Contains(exePath, "/homebrew/")
}

// installBinary replaces the current binary with the new one.
func installBinary(newBinaryPath, targetPath string) error {
	// Check if we can write to the target directory
	targetDir := filepath.Dir(targetPath)
	if err := checkWritePermission(targetDir); err != nil {
		return fmt.Errorf("insufficient permissions for %s: %w\nTry running with sudo", targetDir, err)
	}

	// Create backup
	backupPath := targetPath + ".bak"
	if err := os.Rename(targetPath, backupPath); err != nil {
		return fmt.Errorf("create backup: %w", err)
	}

	// Copy new binary to target (can't rename across filesystems)
	if err := copyFile(newBinaryPath, targetPath); err != nil {
		// Restore backup on failure
		if restoreErr := os.Rename(backupPath, targetPath); restoreErr != nil {
			return fmt.Errorf("install failed (%w) and restore failed (%v)", err, restoreErr)
		}
		return fmt.Errorf("install failed, restored backup: %w", err)
	}

	// Set executable permissions
	if err := os.Chmod(targetPath, 0755); err != nil {
		// Restore backup on failure
		_ = os.Remove(targetPath)
		if restoreErr := os.Rename(backupPath, targetPath); restoreErr != nil {
			return fmt.Errorf("chmod failed (%w) and restore failed (%v)", err, restoreErr)
		}
		return fmt.Errorf("chmod failed, restored backup: %w", err)
	}

	// Remove backup on success
	_ = os.Remove(backupPath)

	return nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return dstFile.Sync()
}

// checkWritePermission checks if we can write to a directory.
func checkWritePermission(dir string) error {
	testFile := filepath.Join(dir, ".nylas-update-test")
	f, err := os.Create(testFile)
	if err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Remove(testFile)
}
