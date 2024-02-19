go mod tidy

go mod tidy
gazelle -go_prefix github.com/cristiandonosoc/gochart
gazelle update-repos -from_file=go.mod -to_macro=deps.bzl%go_dependencies -prune
gazelle update-repos -to_macro=deps.bzl%go_dependencies github.com/cristiandonosoc/golib
gazelle update-repos -to_macro=deps.bzl%go_dependencies github.com/cristiandonosoc/gunreal
gazelle update
