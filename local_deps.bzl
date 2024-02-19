load("@bazel_gazelle//:deps.bzl", "go_repository")

# These are local dependencies that we can activate on and off.
# NOTE(cdc): This file is for local development, so the paths are hardcoded to where they are on my
#            machine. You will need to update them accordingly.

def go_local_dependencies():
    pass
    # go_repository(
    #     name = "com_github_cristiandonosoc_golib",
    #     importpath = "github.com/cristiandonosoc/golib",
    #     vcs = "git",
    #     remote = "/home/cdc/projects/golib",
    #     commit = "0e389513eeb4d7a12998a01f7e86744ef33244f7",
    # )
    # go_repository(
    #     name = "com_github_cristiandonosoc_gunreal",
    #     importpath = "github.com/cristiandonosoc/gunreal",
    #     vcs = "git",
    #     remote = "D:\\projects\\go\\gunreal",
    #     commit = "5f38dcaab0b7dc072f31587b86ccc407e0e53596",
    # )
