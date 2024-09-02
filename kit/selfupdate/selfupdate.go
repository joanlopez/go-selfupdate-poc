package selfupdate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/joanlopez/go-selfupdate-poc/kit/osext"
	"github.com/joanlopez/go-selfupdate-poc/kit/semver"
)

var (
	ErrInvalidSlug        = fmt.Errorf("invalid slug format, it must be owner/name")
	ErrRepositoryNotFound = fmt.Errorf("repository or release not found")
	ErrReleaseNotDetected = fmt.Errorf("release not detected")
	ErrDownload           = fmt.Errorf("release could not be downloaded")
	ErrChecksumValidation = fmt.Errorf("checksum validation failed")
	ErrChecksumDownload   = fmt.Errorf("checksum could not be downloaded")
	ErrDecompression      = fmt.Errorf("release could not be decompressed")
	ErrReleaseBinary      = fmt.Errorf("release archive does not contain the binary")
)

// From checks if there is a new version available and updates the binary.
// It returns true if the binary was updated, false if it is already up-to-date.
// If the binary was updated, the caller should exit immediately.
func From(ctx context.Context, current semver.Version, slug string) (bool, *Release, error) {
	// First, we try to get the path of the current executable.
	// Which, later, will be used to replace the binary.
	cmdPath, err := osext.Executable()
	if err != nil {
		return false, nil, fmt.Errorf("failed to get executable path: %s", err)
	}

	// When on Windows, the executable path might have the '.exe' suffix.
	if runtime.GOOS == "windows" && !strings.HasSuffix(cmdPath, ".exe") {
		cmdPath = cmdPath + ".exe"
	}

	// Check if the binary is a symlink.
	stat, err := os.Lstat(cmdPath)
	if err != nil {
		return false, nil, fmt.Errorf("failed to stat: %s - file may not exist: %s", cmdPath, err)
	}

	// If it is, we resolve the symlink.
	if stat.Mode()&os.ModeSymlink != 0 {
		p, err := filepath.EvalSymlinks(cmdPath)
		if err != nil {
			return false, nil, fmt.Errorf("failed to resolve symlink: %s - for executable: %s", cmdPath, err)
		}
		cmdPath = p
	}

	// Then, we try to detect the latest release.
	// If the current version is the latest, we return early.
	rel, ok, err := detectLatest(ctx, slug)
	if err != nil {
		return false, nil, err
	}
	if !ok {
		return false, nil, ErrReleaseNotDetected
	}

	if current.Equals(rel.Version) {
		return false, rel, nil
	}

	// If not, we update the binary.
	if err := updateTo(ctx, rel, cmdPath); err != nil {
		return false, rel, err
	}

	return true, rel, nil
}

// Release represents a release asset for current OS and arch.
type Release struct {
	// Version is the version of the release
	Version semver.Version
	// AssetURL is a URL to the uploaded file for the release
	AssetURL string
	// AssetSize represents the size of asset in bytes
	AssetByteSize int
	// AssetID is the ID of the asset on GitHub
	AssetID int64
	// ValidationAssetID is the ID of additional validation asset on GitHub
	ValidationAssetID int64
	// URL is a URL to release page for browsing
	URL string
	// ReleaseNotes is a release notes of the release
	ReleaseNotes string
	// Name represents a name of the release
	Name string
	// PublishedAt is the time when the release was published
	PublishedAt *time.Time
	// RepoOwner is the owner of the repository of the release
	RepoOwner string
	// RepoName is the name of the repository of the release
	RepoName string
}
