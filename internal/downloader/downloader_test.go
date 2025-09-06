package downloader

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
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
			_, err := io.WriteString(w, "binary data")
			require.NoError(t, err)
		}))
		defer server.Close()

		tmpDir := t.TempDir()
		downloader, err := New()
		require.NoError(t, err)
		downloader.WithTmpFolder(tmpDir).WithInstallFolder(tmpDir)

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
		tmpFolder := t.TempDir()
		downloader.WithTmpFolder(tmpFolder).WithInstallFolder(tmpFolder)
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
			_, err := io.WriteString(w, "binary data")
			require.NoError(t, err)
		}))
		defer server.Close()

		tmpFolder := t.TempDir()
		downloader, err := New()
		require.NoError(t, err)
		downloader.WithTmpFolder(tmpFolder).WithInstallFolder(tmpFolder)

		binaryInfo := &dto.BinaryInfo{Name: "tool", Owner: "user", InstalledVersion: "v1.0.0"}
		file, err := downloader.downloadToTmpDir(binaryInfo, server.URL+"/foo")

		require.NoError(t, err)
		filePath := filepath.Join(tmpFolder, "azabox-foo")
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
			_, err := io.WriteString(w, "binary data")
			require.NoError(t, err)
		}))
		defer server.Close()

		downloader, err := New()
		require.NoError(t, err)
		downloader.WithTmpFolder(tmpDir)

		err = os.Chmod(tmpDir, 0o400)
		require.NoError(t, err)

		binaryInfo := &dto.BinaryInfo{Name: "tool", Owner: "user", InstalledVersion: "v1.0.0"}
		file, err := downloader.downloadToTmpDir(binaryInfo, server.URL+"/foo")

		assert.Error(t, err)
		assert.Empty(t, file)
	})
}

func TestDownloader_InstallBinary(t *testing.T) {
	t.Run("should install binary", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		tmpDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmpDir)
		t.Setenv("HOME", tmpDir)

		initLogger()
		downloader, err := New()
		require.NoError(t, err)
		downloader.WithInstallFolder(tmpDir)
		tmpFile := filepath.Join(t.TempDir(), "dummy")
		_ = os.WriteFile(tmpFile, []byte("dummy"), 0o600)

		bin := &dto.BinaryInfo{Name: "dummy", InstalledVersion: "v1.0.0"}
		outPath, err := downloader.installBinary(bin, tmpFile)
		require.NoError(t, err)

		info, err := os.Stat(outPath)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(tmpDir, "dummy-v1.0.0"), outPath)
		assert.True(t, info.Mode()&0o100 != 0, "should have exec right")
		assert.False(t, info.IsDir(), "it should not be a directory")
	})

	t.Run("should install binary from archive", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		tmpDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmpDir)
		t.Setenv("HOME", tmpDir)

		initLogger()
		downloader, err := New()
		require.NoError(t, err)
		downloader.WithInstallFolder(tmpDir)
		zipFile := filepath.Join(t.TempDir(), "dummy.zip")

		out, err := os.Create(zipFile)
		require.NoError(t, err)

		zipWriter := zip.NewWriter(out)
		w, err := zipWriter.Create("dummy")
		require.NoError(t, err)

		_, err = w.Write([]byte("dummy"))
		require.NoError(t, err)
		_ = zipWriter.Close()
		_ = out.Close()

		bin := &dto.BinaryInfo{Name: "dummy", InstalledVersion: "v1.0.0"}
		outPath, err := downloader.installBinary(bin, zipFile)
		require.NoError(t, err)

		info, err := os.Stat(outPath)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(tmpDir, "dummy-v1.0.0"), outPath)
		assert.True(t, info.Mode()&0o100 != 0, "should have exec right")
		assert.False(t, info.IsDir(), "it should not be a directory")
	})

	t.Run("should handle corrupted archive", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		tmpDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmpDir)
		t.Setenv("HOME", tmpDir)

		initLogger()
		downloader, err := New()
		require.NoError(t, err)
		downloader.WithInstallFolder(tmpDir)
		tmpFile := filepath.Join(t.TempDir(), "dummy.zip")
		_ = os.WriteFile(tmpFile, []byte("dummy"), 0o600)

		bin := &dto.BinaryInfo{Name: "dummy", InstalledVersion: "v1.0.0"}
		outPath, err := downloader.installBinary(bin, tmpFile)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a valid zip")
		assert.Empty(t, outPath)
	})

	t.Run("should handle open error", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		tmpDir := t.TempDir()
		initLogger()
		downloader, err := New()
		require.NoError(t, err)
		downloader.WithInstallFolder(tmpDir)
		nonExistentPath := filepath.Join(t.TempDir(), "nonExistent")
		bin := &dto.BinaryInfo{Name: "dummy", InstalledVersion: "v1.0.0"}
		outPath, err := downloader.installBinary(bin, nonExistentPath)
		assert.Error(t, err)
		assert.Empty(t, outPath)
	})

	t.Run("should handle read only folder error", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		tmpDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmpDir)
		t.Setenv("HOME", tmpDir)

		err := os.Chmod(tmpDir, 0o400)
		require.NoError(t, err)
		initLogger()
		downloader, err := New()
		require.NoError(t, err)
		downloader.WithInstallFolder(tmpDir)
		tmpFile := filepath.Join(t.TempDir(), "dummy")
		_ = os.WriteFile(tmpFile, []byte("dummy"), 0o600)

		bin := &dto.BinaryInfo{Name: "dummy", InstalledVersion: "v1.0.0"}
		outPath, err := downloader.installBinary(bin, tmpFile)
		assert.Error(t, err)
		assert.Empty(t, outPath)
	})
}

func TestExtractArchive(t *testing.T) {
	t.Run("should handle zip archive", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLogger()
		tmpDir := t.TempDir()
		zipFile := filepath.Join(tmpDir, "tool.zip")
		out, err := os.Create(zipFile)
		require.NoError(t, err)

		zipWriter := zip.NewWriter(out)
		w, err := zipWriter.Create("dummy")
		require.NoError(t, err)

		_, err = w.Write([]byte("dummy"))
		require.NoError(t, err)
		_ = zipWriter.Close()
		_ = out.Close()

		path, err := extractArchive(zipFile, "dummy")
		require.NoError(t, err)
		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.False(t, info.IsDir(), "it should not be a directory")
	})

	t.Run("should handle no binary found zip archive", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLogger()
		tmpDir := t.TempDir()
		zipFile := filepath.Join(tmpDir, "tool.zip")
		out, err := os.Create(zipFile)
		require.NoError(t, err)

		zipWriter := zip.NewWriter(out)
		w, err := zipWriter.Create("foo")
		require.NoError(t, err)

		_, err = w.Write([]byte("dummy"))
		require.NoError(t, err)
		_ = zipWriter.Close()
		_ = out.Close()

		path, err := extractArchive(zipFile, "dummy")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "binary found in zip")
		assert.Empty(t, path)
	})

	t.Run("should handle error on zip archive", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLogger()
		tmpDir := t.TempDir()
		invalidZip := filepath.Join(tmpDir, "invalid.zip")
		_ = os.WriteFile(invalidZip, []byte("dummy"), 0o600)
		path, err := extractZip(invalidZip, "mytool")
		assert.Error(t, err)
		assert.Empty(t, path)
	})

	t.Run("should handle tar gz archive", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLogger()
		tmpDir := t.TempDir()
		tgzFile := filepath.Join(tmpDir, "tool.tar.gz")
		out, err := os.Create(tgzFile)
		require.NoError(t, err)

		gzWriter := gzip.NewWriter(out)
		tarWriter := tar.NewWriter(gzWriter)
		err = tarWriter.WriteHeader(&tar.Header{Name: "dummy", Mode: 0o755, Size: int64(len("dummy"))})
		require.NoError(t, err)

		_, err = tarWriter.Write([]byte("dummy"))
		require.NoError(t, err)

		_ = tarWriter.Close()
		_ = gzWriter.Close()
		_ = out.Close()

		path, err := extractArchive(tgzFile, "dummy")
		require.NoError(t, err)
		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.False(t, info.IsDir(), "it should not be a directory")
	})

	t.Run("should handle no binary found in tar gz archive", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLogger()
		tmpDir := t.TempDir()
		tgzFile := filepath.Join(tmpDir, "tool.tar.gz")
		out, err := os.Create(tgzFile)
		require.NoError(t, err)

		gzWriter := gzip.NewWriter(out)
		tarWriter := tar.NewWriter(gzWriter)
		err = tarWriter.WriteHeader(&tar.Header{Name: "foo", Mode: 0o755, Size: int64(len("dummy"))})
		require.NoError(t, err)

		_, err = tarWriter.Write([]byte("dummy"))
		require.NoError(t, err)

		_ = tarWriter.Close()
		_ = gzWriter.Close()
		_ = out.Close()

		path, err := extractArchive(tgzFile, "dummy")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "binary found in tar.gz")
		assert.Empty(t, path)
	})

	t.Run("should handle error on tar gz archive", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLogger()
		tmpDir := t.TempDir()
		invalidTgz := filepath.Join(tmpDir, "invalid.tar.gz")
		_ = os.WriteFile(invalidTgz, []byte("dummy"), 0o600)
		path, err := extractZip(invalidTgz, "mytool")
		assert.Error(t, err)
		assert.Empty(t, path)
	})
}

func TestIsArchiveFormat(t *testing.T) {
	testCases := []struct {
		name     string
		file     string
		expected bool
	}{
		{
			name:     "should handle zip",
			file:     "foo.zip",
			expected: true,
		},
		{
			name:     "should handle tar.gz",
			file:     "foo.tar.gz",
			expected: true,
		},
		{
			name:     "should handle tgz",
			file:     "foo.tgz",
			expected: true,
		},
		{
			name:     "should handle non archive",
			file:     "foo.exe",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := isArchiveFormat(tc.file)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestDownloader_CreateSymlink(t *testing.T) {
	t.Run("should create symlink", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLogger()
		tmpDir := t.TempDir()
		downloader, err := New()
		require.NoError(t, err)
		downloader.WithInstallFolder(tmpDir)
		binaryInfo := &dto.BinaryInfo{Name: "dummy"}
		target := filepath.Join(t.TempDir(), "dummy")
		_ = os.WriteFile(target, []byte("binary content"), 0o600)

		err = downloader.createSymlink(binaryInfo, target)
		require.NoError(t, err)

		symlink := filepath.Join(tmpDir, binaryInfo.Name)
		_, err = os.Lstat(symlink)
		assert.NoError(t, err, "expected symlink to exist")
	})

	t.Run("should handle error", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLogger()
		tmpDir := t.TempDir()
		downloader, err := New()
		require.NoError(t, err)
		downloader.WithInstallFolder("nonExisting")
		binaryInfo := &dto.BinaryInfo{Name: "dummy"}
		err = downloader.createSymlink(binaryInfo, filepath.Join(tmpDir, "nonExisting"))
		require.Error(t, err)
	})
}
