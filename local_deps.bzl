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
    #     remote = "D:\\projects\\go\\golib",
    #     commit = "e804c8220979431253d13177bce6209bda18a483",
    # )
    # go_repository(
    #     name = "com_github_cristiandonosoc_gunreal",
    #     importpath = "github.com/cristiandonosoc/gunreal",
    #     vcs = "git",
    #     remote = "D:\\projects\\go\\gunreal",
    #     commit = "5f38dcaab0b7dc072f31587b86ccc407e0e53596",
    # )
