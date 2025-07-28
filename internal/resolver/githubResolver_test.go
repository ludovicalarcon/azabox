package resolver

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
