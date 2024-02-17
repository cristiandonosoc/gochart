load("@bazel_gazelle//:deps.bzl", "go_repository")

# These are local dependencies that we can activate on and off.
# NOTE(cdc): This file is for local development, so the paths are hardcoded to where they are on my
#            machine. You will need to update them accordingly.

def go_local_dependencies():
    pass
#    go_repository(
#        name = "com_github_cristiandonosoc_gunreal",
#        importpath = "github.com/cristiandonosoc/gunreal",
#        vcs = "git",
#        remote = "D:\\projects\\go\\gunreal",
#        commit = "80a15b903f6b3c7f4e5bf834cdcb757e95f1c19f",
#    )
