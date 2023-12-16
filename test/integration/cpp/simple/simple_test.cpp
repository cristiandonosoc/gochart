#include "simple.define.h"

#include <gtest/gtest.h>

#include <iostream>
#include <map>
#include <string>

namespace
{

std::string ExtractFunctionName(const std::string& function_name)
{
    size_t pos = function_name.rfind("::");
    if (pos != std::string::npos)
    {
        // +2 to skip "::".
        return function_name.substr(pos + 2);
    }
    // Return the original name if "::" not found.
    return function_name;
}

} // namespace

#define COUNT_CALL() Calls.push_back(ExtractFunctionName(__FUNCTION__));

class Delegate
{
public:
    // Statechart begin.
    void StateDoor_CaptureUpdate(bool tick) { COUNT_CALL(); }
    void StateDoorInactive_CaptureUpdate(bool tick) { COUNT_CALL(); }
    void StateDoorOpened_CaptureUpdate(bool tick) { COUNT_CALL(); }
    void StateDoorMoving_EnterToggleDoor(bool open, bool fast) { COUNT_CALL(); }
    void StateWindowInitial_Enter() { COUNT_CALL(); }
    void StateWindowInitial_Exit() { COUNT_CALL(); }
    void StateWindowIdle_Enter() { COUNT_CALL(); }

    void ActionFunction(bool active) { COUNT_CALL(); }
    void NullTransitionCall() { COUNT_CALL(); }
    // Statechart end.

    const std::vector<std::string>& GetCalls() const { return Calls; }

    size_t GetCallCount(const char* function_name) const
    {
        size_t count = 0;
        for (const auto& call : Calls)
        {
            if (call == function_name)
            {
                count++;
            }
        }

        return count;
    }

private:
    std::vector<std::string> Calls;
};

TEST(SimpleTest, BasicAssertions)
{
    Delegate delegate;
    auto statechart = gochart::Door::Create(&delegate);
    statechart->Activate();
    statechart->TriggerActivateDoorActuator(true);
    statechart->TriggerToggleDoor(true, false);
    statechart->Deactivate();

    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 1);
}

TEST(SimpleTest, Captures)
{
    Delegate delegate;
    auto statechart = gochart::Door::Create(&delegate);

    statechart->Activate();
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::Door));
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::DoorInactive));
    EXPECT_FALSE(statechart->IsInStateExact(gochart::Door::EState::Door));
    EXPECT_TRUE(statechart->IsInStateExact(gochart::Door::EState::DoorInactive));
    EXPECT_FALSE(statechart->IsInState(gochart::Door::EState::DoorActive)); // Check once.
    EXPECT_EQ(delegate.GetCalls().size(), 0);

    statechart->TriggerUpdate(true);
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::Door));
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::DoorInactive));
    EXPECT_TRUE(statechart->IsInStateExact(gochart::Door::EState::DoorInactive));
    EXPECT_EQ(delegate.GetCalls().size(), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 0);

    statechart->TriggerActivateDoorActuator(false);
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::Door));
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::DoorInactive));
    EXPECT_TRUE(statechart->IsInStateExact(gochart::Door::EState::DoorInactive));
    EXPECT_EQ(delegate.GetCalls().size(), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 0);

    statechart->TriggerUpdate(true);
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::Door));
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::DoorInactive));
    EXPECT_TRUE(statechart->IsInStateExact(gochart::Door::EState::DoorInactive));
    EXPECT_EQ(delegate.GetCalls().size(), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 0);

    statechart->TriggerActivateDoorActuator(true);
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::Door));
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::DoorActive));
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::DoorOpened));
    EXPECT_TRUE(statechart->IsInStateExact(gochart::Door::EState::DoorOpened));
    EXPECT_FALSE(statechart->IsInState(gochart::Door::EState::DoorInactive));
    EXPECT_EQ(delegate.GetCalls().size(), 3);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 0);

    statechart->TriggerUpdate(true);
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::Door));
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::DoorActive));
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::DoorOpened));
    EXPECT_TRUE(statechart->IsInStateExact(gochart::Door::EState::DoorOpened));
    EXPECT_EQ(delegate.GetCalls().size(), 4);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 0);

    statechart->TriggerUpdate(false);
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::Door));
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::DoorActive));
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::DoorOpened));
    EXPECT_TRUE(statechart->IsInStateExact(gochart::Door::EState::DoorOpened));
    EXPECT_EQ(delegate.GetCalls().size(), 5);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 0);

    statechart->TriggerToggleDoor(true, false);
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::Door));
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::DoorActive));
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::DoorMoving));
    EXPECT_TRUE(statechart->IsInStateExact(gochart::Door::EState::DoorMoving));
    EXPECT_EQ(delegate.GetCalls().size(), 6);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 1);

    statechart->TriggerUpdate(true);
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::Door));
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::DoorActive));
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::DoorMoving));
    EXPECT_TRUE(statechart->IsInStateExact(gochart::Door::EState::DoorMoving));
    EXPECT_EQ(delegate.GetCalls().size(), 7);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 1);

    statechart->TriggerToggleDoor(true, false);
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::Door));
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::DoorActive));
    EXPECT_TRUE(statechart->IsInState(gochart::Door::EState::DoorMoving));
    EXPECT_TRUE(statechart->IsInStateExact(gochart::Door::EState::DoorMoving));
    EXPECT_EQ(delegate.GetCalls().size(), 7);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 1);

    statechart->Deactivate();
}

TEST(SimpleTest, TransitionAction)
{
    return;
    Delegate delegate;
    auto statechart = gochart::Door::Create(&delegate);

    statechart->Activate();
    EXPECT_EQ(delegate.GetCalls().size(), 0);

    statechart->TriggerActivateDoorActuator(false);
    EXPECT_EQ(delegate.GetCalls().size(), 0);

    statechart->TriggerActivateDoorActuator(true);
    EXPECT_EQ(delegate.GetCalls().size(), 1);
    EXPECT_EQ(delegate.GetCallCount("ActionFunction"), 1);

    statechart->Deactivate();
}

TEST(SimpleTest, NullTransition)
{
    Delegate delegate;
    auto statechart = gochart::Door::Create(&delegate);

    statechart->Activate();

    statechart->TriggerTransferToWindow();
    EXPECT_TRUE(statechart->IsInStateExact(gochart::Door::EState::WindowIdle));

    std::vector<std::string> wantCalls{
        "StateWindowInitial_Enter",
        "NullTransitionCall",
        "StateWindowInitial_Exit",
        "StateWindowIdle_Enter",
    };
    EXPECT_EQ(wantCalls, delegate.GetCalls());

    EXPECT_EQ(delegate.GetCallCount("StateWindowInitial_Enter"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateWindowInitial_Exit"), 1);

    statechart->Deactivate();
}
