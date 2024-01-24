#include "sizeof_parallel.gochart.declare.h"
#include "sizeof_simple.gochart.declare.h"

#include <gtest/gtest.h>

#include <string>
#include <vector>

class Delegate
{
public:
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

    void TrackCall(const std::string& function_name) { Calls.push_back(function_name); }

    std::vector<std::string> Calls;
};

TEST(Sizeof, AssertSize)
{
    using SimpleGochartType = gochart::Simple::Gochart<Delegate>;
    EXPECT_LE(sizeof(SimpleGochartType), 32);

    using ParallelGochartType = gochart::Parallel::Gochart<Delegate>;
    EXPECT_LE(sizeof(ParallelGochartType), 48);
}
