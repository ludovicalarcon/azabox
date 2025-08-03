package downloader

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"gitlab.com/ludovic-alarcon/azabox/internal/dto"
	"gitlab.com/ludovic-alarcon/azabox/internal/logging"
)

type Downloader struct {
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

func New() (*Downloader, error) {
	installFolder, err := getUserLocalBinaryFolder()
	if err != nil {
		return nil, err
	}
	return &Downloader{
		tmpFolder:     os.TempDir(),
		installFolder: installFolder,
	}, nil
}

func (d *Downloader) WithInstallFolder(binaryPath string) *Downloader {
	d.installFolder = binaryPath
	return d
}

func (d *Downloader) WithTmpFolder(tmpPath string) *Downloader {
	d.tmpFolder = tmpPath
	return d
}

func (d *Downloader) Install(binaryInfo *dto.BinaryInfo, url string) error {
	tempFile, err := d.downloadToTmpDir(binaryInfo, url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	fmt.Printf("downloaded to %s\n", tempFile)
	return nil
}

func (d *Downloader) downloadToTmpDir(binaryInfo *dto.BinaryInfo, url string) (string, error) {
	logging.Logger.Debugw("Downloading", "url", url, "binary", binaryInfo.Name, "owner",
		binaryInfo.Owner, "version", binaryInfo.InstalledVersion)
	fmt.Printf("Downloading %s - %s\n", binaryInfo.FullName, binaryInfo.InstalledVersion)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: %s", resp.Status)
	}

	tmpFileName := fmt.Sprintf("azabox-%s", getFileName(url))
	tempFile, err := os.Create(filepath.Join(d.tmpFolder, tmpFileName))
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
