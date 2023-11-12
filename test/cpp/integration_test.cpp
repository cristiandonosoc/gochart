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

class Delegate
{
public:
    // Statechart begin.
    void StateDoorMoving_EnterToggleDoor(bool open)
    {
        CallCountMap[ExtractFunctionName(__FUNCTION__)]++;
    }
    // Statechart end.

    int GetCallCount(const char* function_name) const {
        auto it = CallCountMap.find(function_name);
        return it != CallCountMap.end() ? it->second : 0;
    }

private:
    std::map<std::string, int> CallCountMap;
};

TEST(HelloTest, BasicAssertions)
{
    Delegate delegate;
    auto statechart = gochart::StatechartDoor<Delegate>::Create(&delegate);
    statechart->Activate();
    statechart->TriggerActivateDoorActuator(true);
    statechart->TriggerToggleDoor(true);
    statechart->Deactivate();

    EXPECT_EQ(delegate.GetCallCount("StateDoorMoving_EnterToggleDoor"), 1);
}
