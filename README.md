# Gochart

**This is very much a WIP, so come back later when there is something useful here**

Gochart is statechart code generator for a version of [statechart](https://statecharts.dev/what-is-a-statechart.html), a useful hierarchical state machine definition that permits several interesting patterns that are very useful when defining FSM (Finite State Machines). This is especially useful when defining gameplay logic for games, which tend to be very FSM-oriented. You can see the [original paper](https://www.sciencedirect.com/science/article/pii/0167642387900359), it's quite good.

Statecharts seem to have a long history of usage in the game industry, especially in Ubisoft. Unreal did a [graphical version of something very akin to statecharts](https://docs.unrealengine.com/5.0/en-US/overview-of-state-tree-in-unreal-engine/), albeit not focused to code and more limited in scope.

The goal of this project is to have a standalone tool that can take statechart definitions (mostly in the shape of a custom spec language called "gochart_lang") and output generated code that can be embedded in other applications. The principal usage is to embed it in C++ Unreal co

## Building

### Native Go

If you just want to build Gochart, the native Go toolchain should suffice:

```
go run ./cmd/gochart
```

### Bazel (for C++ and full flow tests)

We use [Bazel](https://bazel.build/) for managing cross language tests, in which we build Gochart,
use it to generate C++ and then run that C++ to ensure it is working as intended as a test.

For building in C++, you will need a correctly set C++ toolchain.
For Windows the easiest way is to install Visual Studio to get the MSVC toolchain. Bazel should be
able to pick it up without (much) configuration.

For building Go with Bazel, the build files are already set, so nothing should be needed.
But if you want to develop on Gochart, it is highly recommended to use [Gazelle](https://github.com/bazelbuild/bazel-gazelle).

Due to compatibility issues with Windows (https://github.com/bazelbuild/bazel-gazelle/issues/1309),
we run Gazelle manually, so you will need to install it:

```
go install github.com/bazelbuild/bazel-gazelle/cmd/gazelle@latest
```

With that, you should be able to run `update_gazelle_deps.bat`, which would be able to track the Go
dependencies within Bazel.
