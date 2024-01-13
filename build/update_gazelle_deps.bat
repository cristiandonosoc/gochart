@set PERCENT=%%

go mod tidy
gazelle -go_prefix github.com/cristiandonosoc/gochart
gazelle update-repos -from_file=go.mod -to_macro=deps.bzl%PERCENT%go_dependencies -prune
