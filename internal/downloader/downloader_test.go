package downloader

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/ludovic-alarcon/azabox/internal/dto"
	"gitlab.com/ludovic-alarcon/azabox/internal/logging"
)

func initLogger() {
	_ = logging.InitLogger(logging.Config{Encoding: logging.Json})
}

func TestDownloader(t *testing.T) {
	t.Run("should create default download", func(t *testing.T) {
		expectedTmpFolder := os.TempDir()
		homeDir, err := os.UserHomeDir()
		require.NoError(t, err)

		expectedUserBinFolder := filepath.Join(homeDir, ".azabox", "bin")
		downloader, err := New()

		require.NoError(t, err)
		assert.Equal(t, downloader.tmpFolder, expectedTmpFolder)
		assert.Equal(t, downloader.installFolder, expectedUserBinFolder)
	})

	t.Run("should override downloader folders", func(t *testing.T) {
		expectedTmpFolder := "foo"
		expectedInstallFolder := "bar"

		downloader, err := New()
		require.NoError(t, err)
		downloader.WithTmpFolder(expectedTmpFolder).WithInstallFolder(expectedInstallFolder)

		assert.Equal(t, downloader.tmpFolder, expectedTmpFolder)
		assert.Equal(t, downloader.installFolder, expectedInstallFolder)
	})

	t.Run("should install", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLogger()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "binary data")
		}))
		defer server.Close()

		tmpDir := t.TempDir()
		downloader, err := New()
		require.NoError(t, err)
		downloader.WithTmpFolder(tmpDir)

		binaryInfo := &dto.BinaryInfo{Name: "tool", Owner: "user", InstalledVersion: "v1.0.0"}
		err = downloader.Install(binaryInfo, server.URL+"/foo")

		require.NoError(t, err)
	})

	t.Run("should handle download error", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLogger()
		server := httptest.NewServer(http.NotFoundHandler())
		defer server.Close()

		downloader, err := New()
		require.NoError(t, err)
		downloader.WithTmpFolder(t.TempDir())
		binaryInfo := &dto.BinaryInfo{Name: "tool", Owner: "user", InstalledVersion: "v1.0.0"}

		err = downloader.Install(binaryInfo, server.URL)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), http.StatusText(http.StatusNotFound))
	})
}

func TestDownloader_DownloadToTmpDir(t *testing.T) {
	t.Run("should download in tmp folder", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLogger()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "binary data")
		}))
		defer server.Close()

		tmpDir := t.TempDir()
		downloader, err := New()
		require.NoError(t, err)
		downloader.WithTmpFolder(tmpDir)

		binaryInfo := &dto.BinaryInfo{Name: "tool", Owner: "user", InstalledVersion: "v1.0.0"}
		file, err := downloader.downloadToTmpDir(binaryInfo, server.URL+"/foo")

		require.NoError(t, err)
		filePath := filepath.Join(tmpDir, "azabox-foo")
		assert.Equal(t, filePath, file)
		info, err := os.Stat(filePath)
		require.NoError(t, err)
		assert.False(t, info.IsDir(), "it should not be a directory")
	})

	t.Run("should handle error", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLogger()
		server := httptest.NewServer(http.NotFoundHandler())
		defer server.Close()

		downloader, err := New()
		require.NoError(t, err)
		downloader.WithTmpFolder(t.TempDir())
		binaryInfo := &dto.BinaryInfo{Name: "tool", Owner: "user", InstalledVersion: "v1.0.0"}

		file, err := downloader.downloadToTmpDir(binaryInfo, server.URL)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), http.StatusText(http.StatusNotFound))
		assert.Empty(t, file)
	})

	t.Run("should handle wrong url", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLogger()
		downloader, err := New()
		require.NoError(t, err)
		downloader.WithTmpFolder(t.TempDir())
		binaryInfo := &dto.BinaryInfo{Name: "tool", Owner: "user", InstalledVersion: "v1.0.0"}

		file, err := downloader.downloadToTmpDir(binaryInfo, "http://%41:8080/")

		assert.Error(t, err)
		assert.Empty(t, file)
	})

	t.Run("should handle error on file creation", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})
		tmpDir := t.TempDir()

		initLogger()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "binary data")
		}))
		defer server.Close()

		downloader, err := New()
		require.NoError(t, err)
		downloader.WithTmpFolder(tmpDir)

		err = os.Chmod(tmpDir, 0400)
		require.NoError(t, err)

		binaryInfo := &dto.BinaryInfo{Name: "tool", Owner: "user", InstalledVersion: "v1.0.0"}
		file, err := downloader.downloadToTmpDir(binaryInfo, server.URL+"/foo")

		assert.Error(t, err)
		assert.Empty(t, file)
	})
}
