package resolver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"gitlab.com/ludovic-alarcon/azabox/internal/dto"
	"gitlab.com/ludovic-alarcon/azabox/internal/logging"
	"gitlab.com/ludovic-alarcon/azabox/internal/platform"
)

const (
	GHBaseAPIUrl                      = "https://api.github.com"
	GHAPIRepoSegment                  = "/repos"
	GHAPIReleaseSegmentTemplate       = "/%s/releases/tags/%s"
	GHAPIReleaseLatestSegmentTemplate = "/%s/releases/%s"

	AcceptHeader    = "application/vnd.github+json"
	UserAgentHeader = "azabox"
)

type GithubResolver struct {
	baseAPIUrl                  string
	reposAPIUrl                 string
	releaseAPIUrlTemplate       string
	releaseLatestAPIUrlTemplate string
}

func NewGithubResolver(baseAPIUrl string) *GithubResolver {
	return &GithubResolver{
		baseAPIUrl:                  baseAPIUrl,
		reposAPIUrl:                 fmt.Sprintf("%s%s", baseAPIUrl, GHAPIRepoSegment),
		releaseAPIUrlTemplate:       fmt.Sprintf("%s%s%s", baseAPIUrl, GHAPIRepoSegment, GHAPIReleaseSegmentTemplate),
		releaseLatestAPIUrlTemplate: fmt.Sprintf("%s%s%s", baseAPIUrl, GHAPIRepoSegment, GHAPIReleaseLatestSegmentTemplate),
	}
}

func createHttpRequest(url string) *http.Request {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", AcceptHeader)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", UserAgentHeader)

	return req
}

func (r *GithubResolver) Resolve(binaryInfo dto.BinaryInfo) (string, error) {
	logging.Logger.Debugw("Resolve binary for github", "binary", binaryInfo.Name, "owner",
		binaryInfo.Owner, "version", binaryInfo.Version)

	url := fmt.Sprintf(r.releaseAPIUrlTemplate, binaryInfo.FullName, binaryInfo.Version)
	if binaryInfo.Version == "latest" {
		url = fmt.Sprintf(r.releaseLatestAPIUrlTemplate, binaryInfo.FullName, binaryInfo.Version)
	}
	req := createHttpRequest(url)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request error: %s", resp.Status)
	}

	var data struct {
		Assets []struct {
			DownloadURL string `json:"browser_download_url"`
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("failed to parse data: %w", err)
	}

	os := runtime.GOOS
	arch := runtime.GOARCH
	archNormalized := platform.NormalizeArch(arch)

	for _, asset := range data.Assets {
		downloadURL := strings.ToLower(asset.DownloadURL)
		if strings.Contains(downloadURL, os) && (strings.Contains(downloadURL, arch) || strings.Contains(downloadURL, archNormalized)) {
			logging.Logger.Debugw("download URL", "url", asset.DownloadURL, "name", binaryInfo.Name,
				"platform", os, "arch", arch, "version", binaryInfo.Version)
			return asset.DownloadURL, nil
		}
	}

	return "", nil
}
