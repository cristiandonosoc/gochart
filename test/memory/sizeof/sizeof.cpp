#include "simple.gochart.define.h"
#include "parallel.gochart.define.h"

#include <format>
#include <iostream>
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

    void TrackCall(const std::string& function_name)
    {
        Calls.push_back(function_name);
    }

    std::vector<std::string> Calls;
};

int main() {
	using SimpleGochartType = gochart::Simple::Gochart<Delegate>;
	std::cout << std::format("Size of generated simple.gochart: {} bytes\n", sizeof(SimpleGochartType));

	using ParallelGochartType = gochart::Parallel::Gochart<Delegate>;
	std::cout << std::format("Size of generated parallel.gochart: {} bytes\n", sizeof(ParallelGochartType));
}
