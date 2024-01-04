#include "simple.define.h"

#include <gtest/gtest.h>

#include <test/test_support.h>

#include <iostream>
#include <map>
#include <string>

class Delegate
{
public:
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
    using GochartType = gochart::Door::Gochart<Delegate>;
    friend GochartType;

    // Gochart begin.
    void StateDoor_CaptureUpdate(bool tick) { TrackCall(__FUNCTION__); }
    void StateDoorInactive_CaptureUpdate(bool tick) { TrackCall(__FUNCTION__); }
    void StateDoorOpened_CaptureUpdate(bool tick) { TrackCall(__FUNCTION__); }
    void StateDoorMoving_EnterToggleDoor(bool open, bool fast) { TrackCall(__FUNCTION__); }
    void StateWindowInitial_Enter() { TrackCall(__FUNCTION__); }
    void StateWindowInitial_Exit() { TrackCall(__FUNCTION__); }
    void StateWindowIdle_Enter() { TrackCall(__FUNCTION__); }

    void ActionFunction(bool active) { TrackCall(__FUNCTION__); }
    void NullTransitionCall() { TrackCall(__FUNCTION__); }
    // Gochart end.

    void TrackCall(const std::string& function_name)
    {
        auto extracted = test::ExtractFunctionName(function_name);

        std::cout << extracted << std::endl;
        Calls.push_back(extracted);
    }

    std::vector<std::string> Calls;
};

TEST(SimpleTest, BasicAssertions)
{
    Delegate delegate;
    auto gochart = gochart::Door::Create(&delegate);
    gochart->Activate();
    gochart->TriggerActivateDoorActuator(true);
    gochart->TriggerToggleDoor(true, false);
    gochart->Deactivate();

    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 1);
}

TEST(SimpleTest, Captures)
{
    Delegate delegate;
    auto gochart = gochart::Door::Create(&delegate);

    gochart->Activate();
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::Door));
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::DoorInactive));
    EXPECT_FALSE(gochart->IsInStateExact(gochart::Door::EState::Door));
    EXPECT_TRUE(gochart->IsInStateExact(gochart::Door::EState::DoorInactive));
    EXPECT_FALSE(gochart->IsInState(gochart::Door::EState::DoorActive)); // Check once.
    EXPECT_EQ(delegate.GetCalls().size(), 0);

    gochart->TriggerUpdate(true);
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::Door));
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::DoorInactive));
    EXPECT_TRUE(gochart->IsInStateExact(gochart::Door::EState::DoorInactive));
    EXPECT_EQ(delegate.GetCalls().size(), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 0);

    gochart->TriggerActivateDoorActuator(false);
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::Door));
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::DoorInactive));
    EXPECT_TRUE(gochart->IsInStateExact(gochart::Door::EState::DoorInactive));
    EXPECT_EQ(delegate.GetCalls().size(), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 0);

    gochart->TriggerUpdate(true);
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::Door));
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::DoorInactive));
    EXPECT_TRUE(gochart->IsInStateExact(gochart::Door::EState::DoorInactive));
    EXPECT_EQ(delegate.GetCalls().size(), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 0);

    gochart->TriggerActivateDoorActuator(true);
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::Door));
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::DoorActive));
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::DoorOpened));
    EXPECT_TRUE(gochart->IsInStateExact(gochart::Door::EState::DoorOpened));
    EXPECT_FALSE(gochart->IsInState(gochart::Door::EState::DoorInactive));
    EXPECT_EQ(delegate.GetCalls().size(), 3);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 0);

    gochart->TriggerUpdate(true);
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::Door));
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::DoorActive));
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::DoorOpened));
    EXPECT_TRUE(gochart->IsInStateExact(gochart::Door::EState::DoorOpened));
    EXPECT_EQ(delegate.GetCalls().size(), 4);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 0);

    gochart->TriggerUpdate(false);
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::Door));
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::DoorActive));
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::DoorOpened));
    EXPECT_TRUE(gochart->IsInStateExact(gochart::Door::EState::DoorOpened));
    EXPECT_EQ(delegate.GetCalls().size(), 5);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 0);

    gochart->TriggerToggleDoor(true, false);
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::Door));
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::DoorActive));
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::DoorMoving));
    EXPECT_TRUE(gochart->IsInStateExact(gochart::Door::EState::DoorMoving));
    EXPECT_EQ(delegate.GetCalls().size(), 6);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 1);

    gochart->TriggerUpdate(true);
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::Door));
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::DoorActive));
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::DoorMoving));
    EXPECT_TRUE(gochart->IsInStateExact(gochart::Door::EState::DoorMoving));
    EXPECT_EQ(delegate.GetCalls().size(), 7);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 1);

    gochart->TriggerToggleDoor(true, false);
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::Door));
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::DoorActive));
    EXPECT_TRUE(gochart->IsInState(gochart::Door::EState::DoorMoving));
    EXPECT_TRUE(gochart->IsInStateExact(gochart::Door::EState::DoorMoving));
    EXPECT_EQ(delegate.GetCalls().size(), 7);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 1);

    gochart->Deactivate();
}

TEST(SimpleTest, TransitionAction)
{
    return;
    Delegate delegate;
    auto gochart = gochart::Door::Create(&delegate);

    gochart->Activate();
    EXPECT_EQ(delegate.GetCalls().size(), 0);

    gochart->TriggerActivateDoorActuator(false);
    EXPECT_EQ(delegate.GetCalls().size(), 0);

    gochart->TriggerActivateDoorActuator(true);
    EXPECT_EQ(delegate.GetCalls().size(), 1);
    EXPECT_EQ(delegate.GetCallCount("ActionFunction"), 1);

    gochart->Deactivate();
}

TEST(SimpleTest, NullTransition)
{
    Delegate delegate;
    auto gochart = gochart::Door::Create(&delegate);

    gochart->Activate();

    gochart->TriggerTransferToWindow();
    EXPECT_TRUE(gochart->IsInStateExact(gochart::Door::EState::WindowIdle));

    std::vector<std::string> wantCalls{
        "StateWindowInitial_Enter",
        "NullTransitionCall",
        "StateWindowInitial_Exit",
        "StateWindowIdle_Enter",
    };
    EXPECT_EQ(wantCalls, delegate.GetCalls());

    EXPECT_EQ(delegate.GetCallCount("StateWindowInitial_Enter"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateWindowInitial_Exit"), 1);

    gochart->Deactivate();
}
