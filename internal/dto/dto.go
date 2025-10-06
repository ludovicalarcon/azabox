package dto

import (
	"fmt"
	"strings"
)

type BinaryInfo struct {
	FullName         string
	Name             string
	Owner            string
	Version          string
	InstalledVersion string
	Resolver         string
}

func (b BinaryInfo) String() string {
	if b.FullName == "" {
		return ""
	}

	if b.Name == b.Owner {
		return fmt.Sprintf("%s in version %s", b.Name, b.InstalledVersion)
	} else {
		return fmt.Sprintf("%s in version %s", b.FullName, b.InstalledVersion)
	}
}

func NormalizeName(name string) string {
	if strings.Contains(name, "/") || name == "" {
		return name
	}
	return fmt.Sprintf("%s/%s", name, name)
}

func (b BinaryInfo) DisplayName() string {
	if b.Name == b.Owner {
		return b.Name
	}
	return b.FullName
}
