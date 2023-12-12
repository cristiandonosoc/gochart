# TODOS

## backend/cpp

- Add option to run clang-format on the generated code.

- Don't generate |CallEnterReaction| if the state has no reactions.
	- We could just detect that case an issue a comment and not generate all that code.

- Investigate if C++20 concepts would create nicer error messages if the owner doesn't correct
  implement the interface.
	- Currently we rely on template generation messages, which are not the greatest.

## CICD

- Setup testing strategy to be run on every commit on github.
  - Should run all the go tests as well as the C++ generated tests.

## Documentation

- [gochart] Generate the Syntax from marked comments in the parser
  - This would be best effort syntax documentation, but should be a good first step.
