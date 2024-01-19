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

Normal usage would be the following:

```
// In header:
    using GochartType = ::gochart::<NAMESPACE>::Gochart;

    // If you're forward declaring.
    std::unique_ptr<GochartType> Gochart;

// In your cpp (normally on init or begin play):
    // We need to pass a pointer to the delegate in the constructor.
    Gochart = std::make_unique<GochartType>(this);
```

For unreal, you can use `TPimplPtr` instead of `std::unique_ptr`. In certain cases in Unrela, the
standard pointer won't be able to work with the pure forwarded type.

[!IMPORTANT]
The gochart instance will have a non-owning pointer to the delegate. It is your responsibility to
ensure that the delegate outlives the gochart instance. The simplest way of achieving this is to
simply have the delegate own the gochart. This also implies that the delegate should be stable in
memory.

If you're using the declare header rather than the forward declare, you can generate generate the
type directly:
```
// In header.
using GochartType = ::gochart::<NAMESPACE>::Gochart;
GochartType gochart;

// In cpp.
gochart.SetOwner(this);
```

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

## States

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

[!TIP]
An enum with all the states present in that gochart can be found in `gochart::<NAMESPACE>::EState`.
You can print their name with `gochart::<NAMESPACE>::ToString(EState)`.

## Triggers

Triggers are the main way users of gochart can interact with the state machine. By sending in
triggers, the gochart will react accordingly, normally in the shape of a [transition](#Transitions)
or a [capture](#Captures).

They are identified by the `trigger` keyword and can only be written at the gochart level:
```
gochart Example {
    trigger Example()
    trigger WithArguments(std::string foo, float dt)
}
```

Triggers will generate functions on the gochart instance that can be used to "send" those triggers
to the gochart:
```
gochart->TriggerExample();
gochart->TriggerWithArguments("foo", 16.6f);
```

[!TIP]
An enum with all the triggers present in the gochart is located in `gochart::<NAMESPACE>::ETrigger`.
You can print their name with `gochart::<NAMESPACE>::ToString(ETrigger)`.

### Trigger Lifetime

Triggers propagate to the gochart and run into completion, running all the appropiate transitions,
captures, reactions, etc. It is fairly common that one of this reactions can send a trigger
themselves. For example:
```
state SomeState {
    // Once SomeTrigger gets called, it could trigger this capture, which will in turn queue the
    // |SomeOtherTrigger| trigger, which will be processed afterwards.
    capture SomeTrigger do { gochart.TriggerSomeOtherTrigger(); }
}
```

Gochart internally has a queue of all the pending triggers waiting to run. Queued triggers are run
in FIFO order. Once the `Trigger*` functions returns, it means that all triggers up to that point
has been run (including the ones that were queued afterwards). At that point, the gochart can be
considered "stable".

### Trigger Arguments

Gochart is able to parse C++ in the arguments and generate the correct signature. If ever other
languages are supported (eg. C#, Go, etc.), Gochart should be able to parse those as well.

An important thing to note is that for a trigger to process, Gochart will copy the argument values
into an internal struct. This means that the arguments _should_ be trivially or cheaply copiable
types. If you need to pass expensive types, you should pass them as pointers:
```
gochart SomeGochart {
    trigger SomeTrigger(MyExpensiveType* foo, const std::vector<ExpensiveStuff>* list)
}
```

[!NOTE]
Because I don't like having references as members, passing references as triggers arguments is
forbidden (Gochart will issue an error). Use pointers instead.

## Conditions/Actions

Most elements of gocharts "react" to some manner to triggers, either by transitioning to other
states, reacting to states being activated/deactivated or simply responding to triggers.

The main mechanism Gochart uses in this case is to call a function on to the delegate that has a
specific name and signature that is depending on the element and trigger. See each specific element
below to see the format of this generated function.

### Inlined Condition

Many elements permit to embed a "condition", which is an inlined C++ function that has to return a
boolean value. The main use case for this

Instead of having Gochart call a generate function, you can inline C++ code directly with the `if`
keyword.
```
enter do { gochart.GetOwner().SomeFunction(); }
```

All inlined functions have a `gochart` variable with is a reference to the gochart instance. You can
use that to trigger things on the statechart or alternatively you can use the `GetOwner()` function
to get a reference to the delegate itself.

In the case of trigger-linked reactions, the trigger parameters will also be available:
```
trigger SomeTrigger(float dt, const std::vector<std::string>* list)

...

exit SomeTrigger do {
    gochart.GetOwner().Update(dt);
#if UE_DEBUG
    debug::Print(*list);
#endif // UE_DEBUG
}
```

[!CAUTION]
Note that in the example above, we are "injecting" a `UE_DEBUG` symbol into the generated code.
Depending on whether you are forward declaring or not, this might not be possible.
See [usage patterns](advanced.md#usage_patters) for more details.



## Reactions

A state can "react" to either being entered, which is when the state is changing from inactive to
active, or being exited, which is changing from active to inactive. Parent states will always enter
before their children, and will always exit after their children. See the [Transition](#transition)
section for more details about how entering and exiting states.

Reactions can have the following shapes:
```
state SomeState {
    // This is a "naked" enter, which will be called always when the state becomes active.
    enter

    // This is a trigger-linked enter, which will only will be called if the state is being entered
    // because that particular trigger was triggered.
    enter SomeTrigger

    // The equivalent can be done for exits:
    exit SomeTrigger
    exit
}
```

In the example above, Gochart will require the delegate to have the following functions, that it
will call when this reactions are happening:

```
// Naked enters generate a function with the format State<STATE_NAME>_Enter:
void StateSomeState_Enter();

// Trigger-linked enters generate a named function State<STATE_NAME>_Enter<TRIGGER_NAME>:
// It will have the arguments parameters of the trigger passed into it.
trigger SomeTrigger(float dt, const std::vector<std::string>* list)

void StateSomeState_EnterSomeTrigger(float dt, const std::vector<std::string>* list);
```
The same happens for exit reactions, except with `Exit` instead of `Enter`.

### Inlined Condition

You can add a condition to the reaction that determines whether the action (callback) will be called.
You can do that by using the `if` keyword, which permits to embed inlined C++, which has to return
a boolean value

```
enter if
```


[!IMPORTANT]



### Inlined Action

## Captures

Captures are similar to reactions that only happen when that particular state is active. If an
active state has a capture for that

