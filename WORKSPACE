load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

register_execution_platforms(
	":x64_windows-clang-cl",
)

register_toolchains(
	"@local_config_cc//:cc-toolchain-x64_windows-clang-cl",
)

# Go -----------------------------------------------------------------------------------------------

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "278b7ff5a826f3dc10f04feaf0b70d48b68748ccd512d7f98bf442077f043fe3",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.41.0/rules_go-v0.41.0.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.41.0/rules_go-v0.41.0.zip",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")
go_rules_dependencies()
go_register_toolchains(version = "1.21.0")

# Gazelle ------------------------------------------------------------------------------------------

http_archive(
    name = "bazel_gazelle",
    sha256 = "29218f8e0cebe583643cbf93cae6f971be8a2484cdcfa1e45057658df8d54002",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.32.0/bazel-gazelle-v0.32.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.32.0/bazel-gazelle-v0.32.0.tar.gz",
    ],
)

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
gazelle_dependencies()

load("//:deps.bzl", "go_dependencies")
# gazelle:repository_macro deps.bzl%go_dependencies
go_dependencies()

# Gtest --------------------------------------------------------------------------------------------

http_archive(
  name = "com_google_googletest",
  urls = ["https://github.com/google/googletest/archive/refs/tags/v1.14.0.zip"],
  strip_prefix = "googletest-1.14.0",
)

# Hedron's Compile Commands Extractor for Bazel ----------------------------------------------------
# https://github.com/hedronvision/bazel-compile-commands-extractor

http_archive(
    name = "hedron_compile_commands",

    # Replace the commit hash in both places (below) with the latest, rather than using the stale one here.
    # Even better, set up Renovate and let it do the work for you (see "Suggestion: Updates" in the README).
    url = "https://github.com/hedronvision/bazel-compile-commands-extractor/archive/e16062717d9b098c3c2ac95717d2b3e661c50608.tar.gz",
    strip_prefix = "bazel-compile-commands-extractor-e16062717d9b098c3c2ac95717d2b3e661c50608",
    # When you first run this tool, it'll recommend a sha256 hash to put here with a message like: "DEBUG: Rule 'hedron_compile_commands' indicated that a canonical reproducible form can be obtained by modifying arguments sha256 = ..."
)
load("@hedron_compile_commands//:workspace_setup.bzl", "hedron_compile_commands_setup")
hedron_compile_commands_setup()

