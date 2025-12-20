package installer

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gitlab.com/ludovic-alarcon/azabox/internal/dto"
	"gitlab.com/ludovic-alarcon/azabox/internal/logging"
)

type Installer interface {
	Install(binaryInfo *dto.BinaryInfo, url string) error
}

type LocalInstaller struct {
	tmpFolder     string
	installFolder string
}

func getUserLocalBinaryFolder() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determined user home directory %w", err)
	}
	return filepath.Join(homeDir, ".azabox", "bin"), nil
}

func getFileName(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return path.Base(parsed.Path)
}

func New() (*LocalInstaller, error) {
	installFolder, err := getUserLocalBinaryFolder()
	if err != nil {
		return nil, err
	}
	return &LocalInstaller{
		tmpFolder:     os.TempDir(),
		installFolder: installFolder,
	}, nil
}

func (l *LocalInstaller) WithInstallFolder(binaryPath string) *LocalInstaller {
	l.installFolder = binaryPath
	return l
}

func (l *LocalInstaller) WithTmpFolder(tmpPath string) *LocalInstaller {
	l.tmpFolder = tmpPath
	return l
}

func (l *LocalInstaller) Install(binaryInfo *dto.BinaryInfo, url string) error {
	tmpFile, err := l.downloadToTmpDir(binaryInfo, url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	targetPath, err := l.installBinary(binaryInfo, tmpFile)
	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}
	if err := l.createSymlink(binaryInfo, targetPath); err != nil {
		return fmt.Errorf("symlink creation failed: %w", err)
	}
	fmt.Println("Installed to " + targetPath)
	return nil
}

func (l *LocalInstaller) downloadToTmpDir(binaryInfo *dto.BinaryInfo, url string) (string, error) {
	logging.Logger().Debug("Downloading", "url", url, "binary", binaryInfo.Name, "owner",
		binaryInfo.Owner, "version", binaryInfo.InstalledVersion)
	fmt.Printf("Downloading %s - %s\n", binaryInfo.FullName, binaryInfo.InstalledVersion)

	resp, err := http.Get(url) //nolint
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: %s", resp.Status)
	}

	tmpFileName := fmt.Sprintf("azabox-%s", getFileName(url))
	tempFile, err := os.Create(filepath.Clean(
		filepath.Join(l.tmpFolder, tmpFileName)))
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

func (l *LocalInstaller) installBinary(binaryInfo *dto.BinaryInfo, tmpPath string) (string, error) {
	if isArchiveFormat(tmpPath) {
		extracted, err := extractArchive(tmpPath, binaryInfo.Name)
		if err != nil {
			return "", err
		}
		tmpPath = extracted
	}

	in, err := os.Open(filepath.Clean(tmpPath))
	if err != nil {
		return "", err
	}
	defer in.Close()

	if err := os.MkdirAll(l.installFolder, 0o750); err != nil {
		return "", err
	}

	targetPath := filepath.Join(l.installFolder, fmt.Sprintf("%s-%s", binaryInfo.Name, binaryInfo.InstalledVersion))
	out, err := os.Create(filepath.Clean(targetPath))
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return "", err
	}
	if err := os.Chmod(targetPath, 0o755); err != nil {
		return "", err
	}
	logging.Logger().Debug("installed binary", "path", targetPath, "binary", binaryInfo.Name,
		"version", binaryInfo.InstalledVersion)
	return targetPath, nil
}

func extractArchive(path string, tool string) (string, error) {
	if strings.HasSuffix(path, ".zip") {
		return extractZip(path, tool)
	} else if strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".tgz") {
		return extractTarGz(path, tool)
	}
	return "", fmt.Errorf("unsupported archive format")
}

func extractZip(path, binaryName string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	tempDir := os.TempDir()
	for _, f := range r.File {
		fileName := f.Name
		if tmp := strings.Split(fileName, "/"); len(tmp) > 1 {
			fileName = tmp[1]
		}
		if strings.Contains(fileName, binaryName) {
			logging.Logger().Debug("copy binary", "name", f.Name)
			outPath := filepath.Join(tempDir, binaryName)
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()
			outFile, err := os.Create(filepath.Clean(outPath))
			if err != nil {
				return "", err
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, rc); err != nil { // nolint
				return "", err
			}
			return outPath, nil
		}
	}
	return "", fmt.Errorf("no matching binary found in zip")
}

func extractTarGz(path, tool string) (string, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	tarReader := tar.NewReader(gz)
	tempDir := os.TempDir()
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		fileName := hdr.Name
		if tmp := strings.Split(fileName, "/"); len(tmp) > 1 {
			fileName = tmp[1]
		}
		if strings.Contains(fileName, tool) {
			logging.Logger().Debug("copy binary", "name", hdr.Name)
			outPath := filepath.Join(tempDir, tool)
			outFile, err := os.Create(filepath.Clean(outPath))
			if err != nil {
				return "", err
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil { // nolint
				return "", err
			}
			return outPath, nil
		}
	}
	return "", fmt.Errorf("no matching binary found in tar.gz")
}

func (l *LocalInstaller) createSymlink(binaryInfo *dto.BinaryInfo, target string) error {
	symLinkPath := filepath.Join(l.installFolder, binaryInfo.Name)
	logging.Logger().Debug("creating symlink", "path", symLinkPath)
	_ = os.Remove(symLinkPath)
	if err := os.Symlink(target, symLinkPath); err != nil {
		return err
	}
	return nil
}

func isArchiveFormat(file string) bool {
	return strings.HasSuffix(file, ".zip") || strings.HasSuffix(file, ".tar.gz") || strings.HasSuffix(file, ".tgz")
}

func IsSupportedFormat(file string) bool {
	return isArchiveFormat(file) || filepath.Ext(file) == "" || filepath.Ext(file) == ".exe"
}
