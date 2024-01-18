# Gochart Documentation

## About Statecharts

Statecharts seem to have a long history of usage in the game industry, especially in Ubisoft.
On their part Unreal did a [graphical version of something very akin to statecharts](https://docs.unrealengine.com/5.0/en-US/overview-of-state-tree-in-unreal-engine/),
albeit not focused to code and more limited in scope (mostly geared to replace/expand their Behavior
Trees.

You can see the [original paper](https://www.sciencedirect.com/science/article/pii/0167642387900359).
It is relatively simple to read.

## Gochart Definition

Gocharts are hierarchical state machines, meaning that states can have another states within them.
When a particular state is active within statechart, also called for the statechart to be "in that
state", all the parent states of that state are also active.


Gochart use their own language to describe statecharts, aptly called GochartLang:

```
// Gochart definition.
gochart D {
    // Triggers.
    trigger Flick()
    trigger TimerElapsed(float dt)


    enter               // Enter reaction.
    enter Flick         // Enter reaction specific to a trigger.

    // Enter reaction with inlined
    enter TimerElapsed do {
        // Some change in my owner. This is native code.
        Owner.Counter++;
    }
    exit                // Exit reaction.

    // State D has both states E and F as children.
    state D {
        state E {
            // If this state is active, when |TimerElapsed| is triggered, it will transition to F.
            transition TimerElapsed > F

            state G {
                // States can also have enter/exit reactions.
                enter
            }
        }

        state F {
            // Transitions can target non-leaf states.
            transition Flick > E
        }
    }
}
```

## C++ Integration

The Gochart tool takes a .gochart file (with content like the one above) and generates C++ headers
meant so that users can instantiate and interact with statecharts.

In particular, the tool will generate a unique namespace called `gochart::<GOCHART_NAME>`. In the
example above, the namespace would be `gochart::D`. The name of this namespace can be changed with
[options](options.md).

Within that namespace, a class called Gochart will be created with the following signature:
```
template <typename Delegate>
class Gochart;
```

The `delegate` type is meant for a class that will hold the callback definitions that gochart will
call for the different events (reactions, captures, etc.).

### Headers

The Gochart tool will generate 3 headers, normally named for the namespace above.

- `<NAMESPACE>.gochart.forward_declare.h`

This one declares the state and triggers enums, as well as a forward declaration of the class above.
This is useful when we want to embed delegates types into the triggers. This means that we will have
to use a pimpl idiom and rely on the .cpp file to compile in the details.

See [usage patterns](advanced.md#usage_patters) for more details.

- `<NAMESPACE>.gochart.declare.h`

This header declares the generated Gochart class.

- `<NAMESPACE>.gochart.define.h`

This holds the actual logic of the gochart logic. This is meant to be compiled only once, so should
be added to a particular .cpp file, so that it only gets added in one translation unit.

### States

The main difference between gocharts and other FSMs is that states have parenthood, a "branch" of
them is active at any particular time (except in the case of parallel states, but we can skip that
for now).

States can have any number of children states, written as a nested expression:
```
state Parent {
    state Child1 {
        ...
    }

    state Child2 {
        ...
    }

    state Child3 {
        ...
    }
}
```

A gochart is "in" a leaf state (a state with no children), and all it's parent states are active as
well. This means that on a transition, gocharts change from a leaf state to another leaf state.

In the example above, the gochart would only be able to transition between G and F.

### Triggers

Triggers are the main way users of gochart can interact with the state machine. By sending in
triggers, the gochart will react accordingly, normally in the shape of a [transition](#Transitions)
or a [capture](#Captures).

### Reactions

A state can "react" to either being entered, which is when the state is changing from inactive to
active, or being exited, which is changing from active to inactive. Parent states will always enter
before their children, and will always exit after their children. See the [Transition](#transition)
section for more details.

TODO: Write




