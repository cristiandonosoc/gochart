// Package build helps embed some build information into the binary.
package build

var Build string

const (
	kDevelopmentVersion = "GOCHART-DEVELOPMENT"
)

func BuildVersion() string {
	if Build != "" {
		return Build
	}

	return kDevelopmentVersion
}

func IsDevBuild() bool {
	build := BuildVersion()
	if build == kDevelopmentVersion {
		return true
	}

	return false
}
