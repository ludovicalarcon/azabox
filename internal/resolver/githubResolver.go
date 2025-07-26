package resolver

import (
	"gitlab.com/ludovic-alarcon/azabox/internal/dto"
	"gitlab.com/ludovic-alarcon/azabox/internal/logging"
)

type GithubResolver struct {
	APIUrl string
}

func (r *GithubResolver) Resolve(binaryInfo dto.BinaryInfo) (string, error) {
	logging.Logger.Debugw("Resolve binary for github", "binary", binaryInfo.Name, "owner",
		binaryInfo.Owner, "version", binaryInfo.Version)

	return "", nil
}
