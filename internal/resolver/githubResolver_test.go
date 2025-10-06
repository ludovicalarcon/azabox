package resolver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/ludovic-alarcon/azabox/internal/dto"
	"gitlab.com/ludovic-alarcon/azabox/internal/logging"
)

func initLogger() {
	_ = logging.InitLogger(logging.Config{Encoding: logging.Json})
}

func newTestBinaryInfo() *dto.BinaryInfo {
	return &dto.BinaryInfo{
		Owner:    "foo",
		Name:     "bar",
		Version:  "v1.0.0",
		FullName: "foo/bar",
	}
}

func TestNewGithubResolver(t *testing.T) {
	t.Run("should return a new resolver", func(t *testing.T) {
		testCases := []struct {
			name string
			url  string
		}{
			{
				name: "tc1",
				url:  GHBaseAPIUrl,
			},
			{
				name: "tc2",
				url:  "example.com",
			},
			{
				name: "tc3",
				url:  "https://foo.bar",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				expected := GithubResolver{
					baseAPIUrl:            tc.url,
					reposAPIUrl:           fmt.Sprintf("%s%s", tc.url, GHAPIRepoSegment),
					releaseAPIUrlTemplate: fmt.Sprintf("%s%s%s", tc.url, GHAPIRepoSegment, GHAPIReleaseSegmentTemplate),
				}
				got := NewGithubResolver(tc.url)

				assert.Equal(t, expected.baseAPIUrl, got.baseAPIUrl)
				assert.Equal(t, expected.reposAPIUrl, got.reposAPIUrl)
				assert.Equal(t, expected.releaseAPIUrlTemplate, got.releaseAPIUrlTemplate)
			})
		}
	})
}

func TestCreateHttpRequest(t *testing.T) {
	t.Run("should properly setup the http request", func(t *testing.T) {
		url := "https://example.com"
		request := createHttpRequest(url)
		assert.Equal(t, url, request.URL.String())
		assert.Equal(t, []string{AcceptHeader}, request.Header["Accept"])
		assert.Equal(t, []string{UserAgentHeader}, request.Header["User-Agent"])
		assert.NotEmpty(t, request.Header[http.CanonicalHeaderKey("X-GitHub-Api-Version")])
	})
}

func TestResolve(t *testing.T) {
	t.Run("should resolve successfully", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		expectedURL := fmt.Sprintf("https://github.com/foo/bar/releases/download/v1.0.0/bar-%s-%s",
			runtime.GOOS, runtime.GOARCH)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := GitHubReleaseResponse{
				Name: "v1.0.0",
				Assets: []GitHubReleaseResponseAsset{
					{
						Url: expectedURL,
					},
				},
			}
			data, err := json.Marshal(resp)
			if err != nil {
				require.NoError(t, err)
			}
			_, err = w.Write(data)
			assert.NoError(t, err)
		}))
		defer server.Close()

		initLogger()
		resolver := NewGithubResolver(server.URL)
		url, err := resolver.Resolve(newTestBinaryInfo())

		assert.NoError(t, err)
		assert.Equal(t, expectedURL, url)
	})

	t.Run("should resolve latest version", func(t *testing.T) {
		testCases := []struct {
			name       string
			extensions string
			empty      bool
		}{
			{
				name:       "tgz",
				extensions: ".tgz",
				empty:      false,
			},
			{
				name:       "tar gz",
				extensions: ".tar.gz",
				empty:      false,
			},
			{
				name:       "zip",
				extensions: ".zip",
				empty:      false,
			},
			{
				name:       "exe",
				extensions: ".exe",
				empty:      false,
			},
			{
				name:       "no extension",
				extensions: "",
				empty:      false,
			},
			{
				name:       "deb",
				extensions: ".deb",
				empty:      true,
			},
			{
				name:       "rpm",
				extensions: ".rpm",
				empty:      true,
			},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Cleanup(func() {
					logging.LogLevel = ""
					logging.Logger = nil
				})

				initLogger()
				binaryInfo := newTestBinaryInfo()
				binaryInfo.Version = LatestVersion
				expectedURL := fmt.Sprintf("https://github.com/foo/bar/releases/download/v1.0.0/bar-%s-%s%s",
					runtime.GOOS, runtime.GOARCH, tc.extensions)
				latestSegmentRelease := fmt.Sprintf(GHAPIReleaseLatestSegmentTemplate, binaryInfo.FullName, binaryInfo.Version)

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.Contains(r.URL.Path, latestSegmentRelease) {
						resp := GitHubReleaseResponse{
							Name: "v1.0.0",
							Assets: []GitHubReleaseResponseAsset{
								{
									Url: expectedURL,
								},
							},
						}
						data, err := json.Marshal(resp)
						if err != nil {
							require.NoError(t, err)
						}
						_, err = w.Write(data)
						require.NoError(t, err)
					} else {
						http.Error(w, "wrong path", http.StatusBadRequest)
					}
				}))
				defer server.Close()

				resolver := NewGithubResolver(server.URL)
				url, err := resolver.Resolve(binaryInfo)

				assert.NoError(t, err)
				if tc.empty {
					assert.Empty(t, url)
				} else {
					assert.Equal(t, expectedURL, url)
				}
			})
		}
	})

	t.Run("should handle no match", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		expectedURL := fmt.Sprintf("https://github.com/foo/bar/releases/download/v1.0.0/bar-%s-%s",
			runtime.GOOS, "arm4242")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := GitHubReleaseResponse{
				Name: "v1.0.0",
				Assets: []GitHubReleaseResponseAsset{
					{
						Url: expectedURL,
					},
				},
			}
			data, err := json.Marshal(resp)
			if err != nil {
				require.NoError(t, err)
			}
			_, err = w.Write(data)
			require.NoError(t, err)
		}))
		defer server.Close()

		initLogger()
		resolver := NewGithubResolver(server.URL)
		url, err := resolver.Resolve(newTestBinaryInfo())

		assert.NoError(t, err)
		assert.Empty(t, url)
	})

	t.Run("should handle request error", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLogger()
		server := httptest.NewServer(http.NotFoundHandler())
		defer server.Close()

		resolver := NewGithubResolver(server.URL)
		url, err := resolver.Resolve(newTestBinaryInfo())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), http.StatusText(http.StatusNotFound))
		assert.Empty(t, url)
	})

	t.Run("should handle invalid json answer", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLogger()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := io.WriteString(w, "invalid-json")
			require.NoError(t, err)
		}))
		defer server.Close()

		resolver := NewGithubResolver(server.URL)
		url, err := resolver.Resolve(newTestBinaryInfo())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse data")
		assert.Empty(t, url)
	})
}

func TestResolveLatestVersion(t *testing.T) {
	t.Run("should resolve latest url", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		binaryInfo := newTestBinaryInfo()
		binaryInfo.Version = LatestVersion
		expectedVersion := "vX.Y.Z"
		URL := fmt.Sprintf("https://github.com/foo/bar/releases/download/v1.0.0/bar-%s-%s%s",
			runtime.GOOS, runtime.GOARCH, ".tar.gz")
		latestSegmentRelease := fmt.Sprintf(GHAPIReleaseLatestSegmentTemplate, binaryInfo.FullName, binaryInfo.Version)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, latestSegmentRelease) {
				resp := GitHubReleaseResponse{
					Name: expectedVersion,
					Assets: []GitHubReleaseResponseAsset{
						{
							Url: URL,
						},
					},
				}
				data, err := json.Marshal(resp)
				if err != nil {
					require.NoError(t, err)
				}
				_, err = w.Write(data)
				require.NoError(t, err)
			} else {
				http.Error(w, "wrong path", http.StatusBadRequest)
			}
		}))
		defer server.Close()

		initLogger()
		resolver := NewGithubResolver(server.URL)
		version, err := resolver.ResolveLatestVersion(*binaryInfo)

		assert.NoError(t, err)
		assert.Equal(t, expectedVersion, version)
	})
}

func TestName(t *testing.T) {
	got := NewGithubResolver("").Name()
	assert.Equal(t, GithubResolverName, got)
}
