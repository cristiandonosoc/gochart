# Advanced Topics

This document is a collection of advanced topics associated with Gocharts. This assumes familiarity
with the basic notions described in the [base documentation](gochart.md).

## Usage Patterns

The Gochart tool generates 3 headers: forward_declare, declare and define. You can see an example of
the generated code [here](example_generated/).

### Forward Declare Header

This just declares trigger and states enums, as well as the forward declaration of the class. This
is useful when we want to use symbols in our gochart that have not been defined yet. For example, if
we have the following class.

```
// The declaration of the gochart comes from here.
#include <simple.gochart.forward_declare.h>

class Delegate {
public:
    // The enum is declared here.
    enum class ESomeEnum : uint8 {
        Value1,
        Value2,
    };

private:
    // We cannot use the declared version as it would need to know about ESomeEnum
    std::unique_ptr<gochart::Simple::Gochart> Gochart;
};
```

If we want our Gochart to use `ESomeEnum` in some trigger or some inlined function, could not use
the declare header, as it would be trying to use symbols that are not declared yet (chicken and egg
situation).

So what we can do is use the forward declaration and have a pointer, which will rely on the
implementation being compiled later ([pimpl idiom](https://en.cppreference.com/w/cpp/language/pimpl)).
Then on the .cpp file, when the define header is used, all the symbols will be correctly defined and
it will work.

The drawback of this approach as we would using one allocation for getting the instance at runtime.
Depending on your use case, in many cases this is fine.

### Declare Header

The declare header goes a step further from the forward declare (it actually includes it) and
declares the whole Gochart class. This shows while we couldn't use the enum in the trigger, because
it would be need to be present in the trigger API, hence needing those symbols to be defined.

Now, if you are not using symbols that are not defined (eg. Your triggers only use base types), then
you can actually rely on having the class and you can avoid the allocation. This can be useful when
you know that you will have a lot of instances of the gochart.

```
#include <simple.gochart.declare.h>

class Delegate {
    ...

    // We use the member directly, since the compiler will know the size. No allocation needed.
    ::gochart::Simple::Gochart Gochart;
};
```

### Define Header

The define header includes the declare header and has all the information needed to have a
functional runtime of the generated gochart. The only reason it is a header is because the Gochart
class is templated, but nominally it should be thought of an implementation associated **with an
unique translation unit**.

This header holds the runtime and a huge amount of switch statements which hold the specific
information about the generated gochart (which state is parent of which, which state has enter
reactions, etc.). The runtime uses this function to react accordingly. Those functions are the main
purpose of the generated code.

> [!IMPORTANT]
> The define header has to be compiled **only once**. Other wise you will start getting duplicated
> symbols. The easiest way is to just add the header into the .cpp class of the class that is
> working as the delegate of the gochart.

## Parallel

Parallel gocharts are a way to define multiple regions (parallel groups in Gochart parlance) within
the Gochart, which can be enabled or not. If a region is enabled, it must have an active state.

> [!TIP]
> It can help to consider that **technically** all Gocharts are parallel, only that simple ones are
> the simplest case of one parallel region. The parallel option just permits to define more regions.

In the following example we describe a gochart that have more parallel regions by using the
`parallel` option. See the [options docs](options.md) for more information on options.

```
gochart MyGochart(parallel) {
    trigger Toggle1()
    trigger Toggle2()

    state ParallelRegion1 {
        state A {
            transition Toggle1 > B
        }

        state B {
            transition Toggle2 > A
        }
    }

    state ParallelRegion2 {
        state C {
            transition Toggle2 > D
        }

        state D {
            ...
        }
    }
}
```

We see now that we have parallel regions: `ParallelRegion1` and `ParallelRegion2`.

> [!TIP]
> **Technically** there are 3 regions, the two defined and the base region that encompasses the
> whole gochart (internally called `__Root__`). This is the base parallel region that all simple
> gochart reside on.

At any given moment, all **enabled** parallel regions must have an active state (more on disabled
parallel regions on [nested parallel regions](#nested-parallel-regions)). In this example, upon
activating the gochart, it will be both in states `A` and `C`.

A property of parallel regions is that they all receive the same trigger. So just issuing one
trigger will make all active parallel regions to react accordingly.

> [!IMPORTANT]
> While all the regions will be receive the trigger, they are evaluated serially, started from the
> top of the file. Meaning that `ParallelRegion1` will evaluate the trigger before `ParallelRegion2`.
> In general it is considered a bad practice to rely on this property though, as parallel states
> should be relatively decoupled.

We can see the effect of issuing triggers with the following snippet:

```
gochart.Activate();
Print(gochart.GetCurrentStates());      // Prints [A, C].

// When issuing a single trigger, both parallel regions will receive it.
gochart.TriggerToggle1();
Print(gochart.GetCurrentStates());      // Prints [B, C].

// We can then issue other triggers.
gochart.TriggerToggle2();
Print(gochart.GetCurrentStates());      // Prints [A, D].
```

### Transitions within parallel regions

States can only issue transitions **within** their parallel region, never outside of it. This is
true even for the parent of a nested parallel region. In the example above, any state under
`ParallelRegion1` cannot issue a transition outside of that state (it will fail to compile). This
means that state `A` can only transition to `B`. By the same rule, `C` can only transition to `D`.

### Nested Parallel Regions

> [!CAUTION]
> It is **rare** to need this case. Consider first if you could write the same gochart in a more
> simple manner.

Normally Gocharts are made parallel on at the root (gochart) level. But it can also be made parallel
at the state level. Up to the point that even a parallel region can have parallel regions of its
own. We extend the previous example to exemplify this:

```
gochart MyGochart(parallel) {
    trigger Toggle1()
    trigger Toggle2()
    trigger ToggleB()

    state ParallelRegion1 {
        state A {
            transition Toggle1 > B
        }

        state B {
            transition Toggle2 > A

            state BChild1(parallel, default) {
                state BChild1_Parallel1 {
                    ...
                }

                state BChild1_Parallel2 {
                    ...
                }
            }
        }
    }

    state ParallelRegion2 {
        state C {
            transition Toggle2 > D
        }

        state D {
            ...
        }
    }
}
```

In the following example, now the statechart has 4 parallel regions: `ParallelRegion1`,
`ParallelRegion2`, `BChild1_Parallel1` and `BChild1_Parallel2`. But when the gochart begins, the
latter two are disabled, since their parent is not active. **Regions are only active if their parent
state is active** (this is always true for the gochart-root state).

Now if we if we issue the same triggers as above, we should get a different result:
```
gochart.Activate();
Print(gochart.GetCurrentStates());      // Prints [A, C, <inactive>, <inactive>].

// When issuing a single trigger, both parallel regions will receive it.
gochart.TriggerToggle1();
Print(gochart.GetCurrentStates());      // Prints [B, C, BChild1_Parallel1, BChildParallel2].

// We can then issue other triggers.
gochart.TriggerToggle2();
Print(gochart.GetCurrentStates());      // Prints [A, D, <inactive>, <inactive>].
```

Note that when we send the `Toggle2` trigger, the gochart exits state `B`, meaning that the parallel
regions below it gets exited as well.

> [!IMPORTANT]
> Parallel regions propagate triggers and exits from *more specific to less specific*. So while it
> does respect the top to bottom in file approach, a more specific (nested) state will evaluate the
> trigger first. In this example, the parallel regions under B will receive the trigger before their
> parent region. Again, it is bad practice to rely on this fact.

> [!TIP]
> See [the parallel test to see a working example of nested parallel behavior](/test/integration/cpp/parallel).
