# TODOS

## frontend/gochart_lang

- Set up better error messaging, taking into account line:char from token.
- Set up an internal error type, that prompts user to file a bug issue.

- Support capture

## backend/cpp

- Move the generated impl into the template so that we don't need to generate the callback struct,
  which is unnecessary data that needs to be tracked per statechart instance.
	- This is also important so we can avoid having to declare all the headers in the statechart.

- Ensure we have a version of std::function that will never allocate.
	- We might need to bring in one.

- Add option to run clang-format on the generated code.

- Don't generate |CallEnterReaction| if the state has no reactions.
	- We could just detect that case an issue a comment and not generate all that code.

- Investigate if C++20 concepts would create nicer error messages if the owner doesn't correct
  implement the interface.
	- Currently we rely on template generation messages, which are not the greatest.

## Testing

- Setup some basic infra to be able to test multiple cases, especially to be able to test wrong
  syntax and evaluate the error messages.

## CICD

- Setup testing strategy to be run on every commit on github.
  - Should run all the go tests as well as the C++ generated tests.

## Documentation

- [gochart] Generate the Syntax from marked comments in the parser
  - This would be best effort syntax documentation, but should be a good first step.

