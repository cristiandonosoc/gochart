# TODOS

## frontend/gochart_lang

- [EPIC] Generate valid IR from lang.

## backend/cpp

- Support default exit reactions
- Support trigger... triggering
- Support trigger arguments
- Support triggered enter reactions
- Support triggered exit reactions

- Ensure we have a version of std::function that will never allocate.
	- We might need to bring in one.
- Add option to run clang-format on the generated code.
- Don't generate |CallEnterReaction| if the state has no reactions.
	- We could just detect that case an issue a comment and not generate all that code.
- Investigate if C++20 concepts would create nicer error messages if the owner doesn't correct
  implement the interface.
	- Currently we rely on template generation messages, which are not the greatest.
