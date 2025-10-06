package resolver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"gitlab.com/ludovic-alarcon/azabox/internal/dto"
	"gitlab.com/ludovic-alarcon/azabox/internal/installer"
	"gitlab.com/ludovic-alarcon/azabox/internal/logging"
	"gitlab.com/ludovic-alarcon/azabox/internal/platform"
)

const (
	GHBaseAPIUrl                      = "https://api.github.com"
	GHAPIRepoSegment                  = "/repos"
	GHAPIReleaseSegmentTemplate       = "/%s/releases/tags/%s"
	GHAPIReleaseLatestSegmentTemplate = "/%s/releases/%s"
	GithubResolverName                = "github"

	AcceptHeader    = "application/vnd.github+json"
	UserAgentHeader = "azabox"
)

type GithubResolver struct {
	baseAPIUrl                  string
	reposAPIUrl                 string
	releaseAPIUrlTemplate       string
	releaseLatestAPIUrlTemplate string
}

type GitHubReleaseResponseAsset struct {
	Url string `json:"browser_download_url"`
}

type GitHubReleaseResponse struct {
	Name   string                       `json:"name"`
	Assets []GitHubReleaseResponseAsset `json:"assets"`
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

func (r GithubResolver) callGithubReleaseEndpoint(binaryInfo dto.BinaryInfo) (GitHubReleaseResponse, error) {
	url := fmt.Sprintf(r.releaseAPIUrlTemplate, binaryInfo.FullName, binaryInfo.Version)
	if binaryInfo.Version == "latest" {
		url = fmt.Sprintf(r.releaseLatestAPIUrlTemplate, binaryInfo.FullName, binaryInfo.Version)
	}
	req := createHttpRequest(url)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return GitHubReleaseResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GitHubReleaseResponse{}, fmt.Errorf("request error: %s", resp.Status)
	}

	var data GitHubReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return GitHubReleaseResponse{}, fmt.Errorf("failed to parse data: %w", err)
	}

	return data, nil
}

func (r GithubResolver) Resolve(binaryInfo *dto.BinaryInfo) (string, error) {
	logging.Logger.Debugw("Resolve binary for github", "binary", binaryInfo.Name, "owner",
		binaryInfo.Owner, "version", binaryInfo.Version)

	data, err := r.callGithubReleaseEndpoint(*binaryInfo)
	if err != nil {
		return "", err
	}

	os := runtime.GOOS
	arch := runtime.GOARCH
	archNormalized := platform.NormalizeArch(arch)
	logging.Logger.Debugw("os info", "os", os, "arch", arch, "normalized", archNormalized)

	for _, asset := range data.Assets {
		downloadURL := strings.ToLower(asset.Url)
		if strings.Contains(downloadURL, os) &&
			(strings.Contains(downloadURL, arch) || strings.Contains(downloadURL, archNormalized)) {
			if installer.IsSupportedFormat(downloadURL) {
				logging.Logger.Debugw("download URL", "url", asset.Url, "name", binaryInfo.Name,
					"platform", os, "arch", arch, "version", binaryInfo.Version, "resolvedVersion", data.Name)
				binaryInfo.InstalledVersion = data.Name
				binaryInfo.Resolver = GithubResolverName
				return asset.Url, nil
			}
		}
	}

	return "", nil
}

func (r GithubResolver) ResolveLatestVersion(binaryInfo dto.BinaryInfo) (string, error) {
	tmpBinaryInfo := binaryInfo
	tmpBinaryInfo.Version = LatestVersion

	data, err := r.callGithubReleaseEndpoint(tmpBinaryInfo)
	return data.Name, err
}

func (r GithubResolver) Name() string {
	return GithubResolverName
}
