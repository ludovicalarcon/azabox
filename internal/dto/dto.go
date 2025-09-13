package dto

import "fmt"

type BinaryInfo struct {
	FullName         string
	Name             string
	Owner            string
	Version          string
	InstalledVersion string
}

func (b *BinaryInfo) String() string {
	if b.FullName == "" {
		return ""
	}

	if b.Name == b.Owner {
		return fmt.Sprintf("%s in version %s", b.Name, b.InstalledVersion)
	} else {
		return fmt.Sprintf("%s in version %s", b.FullName, b.InstalledVersion)
	}
}
