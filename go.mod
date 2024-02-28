module github.com/cristiandonosoc/gochart

go 1.21.5

// Activate these replace statements for local development.
// NOTE(cdc): This file is for local development, so the paths are hardcoded to where they are on my
//             machine. You will need to update them accordingly.
//replace github.com/cristiandonosoc/golib => D:\projects\go\golib
//replace github.com/cristiandonosoc/gunreal => D:\projects\go\gunreal

require (
	github.com/Masterminds/sprig/v3 v3.2.3
	github.com/cristiandonosoc/golib v0.1.6
	github.com/cristiandonosoc/gunreal v0.1.3
	github.com/stretchr/testify v1.8.4
	github.com/urfave/cli/v2 v2.25.7
	golang.org/x/exp v0.0.0-20240222234643-814bf88cf225
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.0 // indirect
	github.com/bazelbuild/rules_go v0.41.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/huandu/xstrings v1.3.3 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/crypto v0.3.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
