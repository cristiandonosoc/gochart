# TODOS

These are not currently sorted by priority or effort size. This is simply a dump of things that we
want to add to Gocharts down the line or things we know we need to do.

## backend/cpp

- Add option to run clang-format on the generated code.

- Don't generate |CallEnterReaction| if the state has no reactions.
	- We could just detect that case an issue a comment and not generate all that code.

- Investigate if C++20 concepts would create nicer error messages if the owner doesn't correct
  implement the interface.
	- Currently we rely on template generation messages, which are not the greatest.

## Documentation

- [gochart] Generate the Syntax from marked comments in the parser
  - This would be best effort syntax documentation, but should be a good first step.

## Logging/Debugging

- Tracing/Logging
    - Currently we don't really have a way to inspect what happened in a gochart.
- Inlined function breakpoint support
    - Being able to Gochart to trigger a breakpoint when an inlined function is going to happen.

## Formatter

Add version string to gochart_fmt, similar to what we do with the main tool.

## Visual Studio Support

- Visual Studio [syntax highlighting](https://code.visualstudio.com/api/language-extensions/syntax-highlight-guide)

## Future Work

- Add per state variables
    - Currently if we want to have data associated with a state, we have to have the owner have
      them, and also track their lifetime (reset on exit, etc.). We could have a way to determine
      variables associated with a state that could be automatically cleared.
