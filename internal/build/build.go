// Package build helps embed some build information into the binary.
package build

var GOCHART_VERSION string
var GOCHART_BUILD_TIME string
var GOCHART_GIT_COMMIT string

const (
	kDevelopmentVersion = "GOCHART-DEVELOPMENT"
)

func BuildVersion() string {
	if GOCHART_VERSION != "" {
		return GOCHART_VERSION
	}

	return kDevelopmentVersion
}

func BuildTime() string {
	if GOCHART_BUILD_TIME != "" {
		return GOCHART_BUILD_TIME
	}

	return "<unspecified>"
}

func GitCommit() string {
	if GOCHART_GIT_COMMIT != "" {
		return GOCHART_GIT_COMMIT
	}

	return "<unspecified>"
}

func IsDevBuild() bool {
	build := BuildVersion()
	if build == kDevelopmentVersion {
		return true
	}

	return false
}
