package platform

func NormalizeArch(goarch string) string {
	switch goarch {
	case "amd64":
		return "x86_64"
	default:
		return goarch
	}
}
