# Gochart

**This is very much a WIP, so come back later when there is something useful here**

Gochart is statechart code generator for a version of [statechart](https://statecharts.dev/what-is-a-statechart.html), a useful hierarchical state machine definition that permits several interesting patterns that are very useful when defining FSM (Finite State Machines). This is especially useful when defining gameplay logic for games, which tend to be very FSM-oriented. You can see the [original paper](https://www.sciencedirect.com/science/article/pii/0167642387900359), it's quite good.

Statecharts seem to have a long history of usage in the game industry, especially in Ubisoft. Unreal did a [graphical version of something very akin to statecharts](https://docs.unrealengine.com/5.0/en-US/overview-of-state-tree-in-unreal-engine/), albeit not focused to code and more limited in scope.

The goal of this project is to have a standalone tool that can take statechart definitions (mostly in the shape of a custom spec language called "gochart_lang") and output generated code that can be embedded in other applications. The principal usage is to embed it in C++ Unreal co



