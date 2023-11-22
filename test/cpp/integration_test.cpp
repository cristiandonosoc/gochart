#include "statechart.generated.h"

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

#define COUNT_CALL()                                                                               \
    CallCountMap[ExtractFunctionName(__FUNCTION__)]++;                                             \
    TotalCalls++;

class Delegate
{
public:
    // Statechart begin.

    void StateDoor_CaptureUpdate(bool tick) { COUNT_CALL(); }
    void StateDoorInactive_CaptureUpdate(bool tick) { COUNT_CALL(); }
    void StateDoorOpened_CaptureUpdate(bool tick) { COUNT_CALL(); }
    void StateDoorMoving_EnterToggleDoor(bool open, bool fast) { COUNT_CALL(); }
    // Statechart end.

    int GetCallCount(const char* function_name) const
    {
        auto it = CallCountMap.find(function_name);
        return it != CallCountMap.end() ? it->second : 0;
    }

    int GetTotalCalls() const { return TotalCalls; }

private:
    std::map<std::string, int> CallCountMap;
    int TotalCalls = 0;
};

TEST(HelloTest, BasicAssertions)
{
    Delegate delegate;
    auto statechart = gochart::StatechartDoor<Delegate>::Create(&delegate);
    statechart->Activate();
    statechart->TriggerActivateDoorActuator(true);
    statechart->TriggerToggleDoor(true, false);
    statechart->Deactivate();

    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 1);
}

TEST(HelloTest, Captures)
{
    Delegate delegate;
    auto statechart = gochart::StatechartDoor<Delegate>::Create(&delegate);

    statechart->Activate();
    EXPECT_EQ(delegate.GetTotalCalls(), 0);

    statechart->TriggerUpdate(true);
    EXPECT_EQ(delegate.GetTotalCalls(), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 0);

    statechart->TriggerActivateDoorActuator(false);
    EXPECT_EQ(delegate.GetTotalCalls(), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 0);

    statechart->TriggerUpdate(true);
    EXPECT_EQ(delegate.GetTotalCalls(), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 0);

    statechart->TriggerActivateDoorActuator(true);
    EXPECT_EQ(delegate.GetTotalCalls(), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 0);

    statechart->TriggerUpdate(true);
    EXPECT_EQ(delegate.GetTotalCalls(), 3);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 0);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 0);

    statechart->TriggerUpdate(false);
    EXPECT_EQ(delegate.GetTotalCalls(), 4);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 0);

    statechart->TriggerToggleDoor(true, false);
    EXPECT_EQ(delegate.GetTotalCalls(), 5);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 1);

    statechart->TriggerUpdate(true);
    EXPECT_EQ(delegate.GetTotalCalls(), 6);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 1);

    statechart->TriggerToggleDoor(true, false);
    EXPECT_EQ(delegate.GetTotalCalls(), 6);
    EXPECT_EQ(delegate.GetCallCount("StateDoor_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorInactive_CaptureUpdate"), 2);
    EXPECT_EQ(delegate.GetCallCount("StateDoorOpened_CaptureUpdate"), 1);
    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 1);

    statechart->Deactivate();
}
