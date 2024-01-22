# Gochart

<p align="center">
    <img src="./extras/images/gochart_logo.png" width="200" />
</p>

Gochart is statechart code generator capable of targeting multiple backends, though currently C++ is
the only one available. While not bound to it at all, it was thought to be used within Unreal as a
way of describing gameplay state machines in a more robust manner.

It is meant to be performance oriented, geared to be used in games, such as the ones made in Unreal:
- Low memory footprint: The generated class is 24 bytes big in the best case, 32 in normal cases.
- No memory allocations: Gocharts do not perform any memory allocations at all.

## Statecharts

[Statecharts](https://statecharts.dev/what-is-a-statechart.html) are an useful hierarchical state
machine definition that permits several interesting patterns that are very useful when defining FSM
(Finite State Machines). This is especially useful when defining gameplay logic for games, which
tend to be very FSM-oriented.

See [Gocharts Doc](docs/gochart.md) for more data.

## Installation & Usage

See [Releases](https://github.com/cristiandonosoc/gochart/releases) for a version for your OS.
You can then use the cli to generate the C++ code as such:

```
gochart backend cpp generate [STATECHART_FILE] [OUTPUT_PATH] [FILE_PREFIX]
```

Example
```
gochart backend cpp generate elevator.gochart Gameplay\ elevator
```
Will generate
```
What would this generate?
```

### Unreal

Gochart also has a special mode for detecting all the gocharts present in an Unreal project:
```
gochart unreal generate [UNREAL_PROJECT_SOURCE_DIR]
```

IMPORTANT: Refer [unreal.md](<PATH TO UNREAL MD>) for the options, as those will determine where
there generated code will be created.

## Building

We use [Bazel](https://bazel.build/) as the main build system for the project, though standard Go
should work as well for compiling the tool.

With Bazel:
```
bazel build //cmd/gochart
```

With native Go:
```
go build cmd/gochart
```

### Gazelle

For building Go with Bazel, the build files are already set, so nothing should be needed.
But if you want to develop on Gochart, it is highly recommended to use
[Gazelle](https://github.com/bazelbuild/bazel-gazelle).

Due to compatibility issues with Windows (https://github.com/bazelbuild/bazel-gazelle/issues/1309),
we run Gazelle manually, so you will need to install it:

```
go install github.com/bazelbuild/bazel-gazelle/cmd/gazelle@latest
```

With that, you should be able to run `update_gazelle_deps.bat`, which would be able to track the Go
dependencies within Bazel.

## Tests

We leverage Bazel for managing cross language tests, in which we build the tool , use it to generate
C++ from a gochart file and then run that C++ to ensure it is working as intended as a test, using
GTest.

For building in C++, you will need a correctly set C++ toolchain. For Windows the easiest way is to
install Visual Studio to get the MSVC toolchain. In Linux, both GCC or Clang should work correctly.

After that, the tests should work correctly:
```
bazel test //...
```
