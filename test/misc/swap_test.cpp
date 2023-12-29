// Simple test to verify that my list inverse function is working.

#include <gtest/gtest.h>

#include <test/test_support.h>

#include <algorithm>
#include <array>
#include <vector>

namespace
{

enum class EState : uint8_t
{
    Invalid,
    State1,
    State2,
    State3,
    State4,
    State5,
    State6,
    State7,
};

const char* ToString(EState state)
{
    switch (state)
    {
    case EState::Invalid:
        return "Invalid";
    case EState::State1:
        return "State1";
    case EState::State2:
        return "State2";
    case EState::State3:
        return "State3";
    case EState::State4:
        return "State4";
    case EState::State5:
        return "State5";
    case EState::State6:
        return "State6";
    case EState::State7:
        return "State7";
    }

    assert(false);
    return "<unknown>";
}

void SwapList(std::vector<EState>& list, size_t first_invalid_index)
{
    std::reverse(list.begin(), list.begin() + first_invalid_index);
}

} // namespace

TEST(ListSwap, MinorCase)
{
    std::vector<EState> list{
        EState::State1,
        EState::State2,
        EState::State3,
        EState::State4,
        EState::Invalid,
        EState::Invalid,
    };

    size_t first_invalid_index = 0;
    for (EState state : list)
    {
        if (state == EState::Invalid)
        {
            break;
        }
        first_invalid_index++;
    }

    SwapList(list, first_invalid_index);

    test::CompareStates<EState>(
        std::array{
            EState::State4,
            EState::State3,
            EState::State2,
            EState::State1,
            EState::Invalid,
            EState::Invalid,
        },
        list);
}
