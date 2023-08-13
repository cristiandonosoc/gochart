@set PERCENT=%%
gazelle update-repos -from_file=go.mod -to_macro=deps.bzl%PERCENT%go_dependencies
