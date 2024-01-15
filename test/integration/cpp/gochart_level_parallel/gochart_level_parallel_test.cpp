#include "gochart_level_parallel.define.h"

#include <test/test_support.h>

#include <gtest/gtest.h>

#include <vector>

class Delegate
{
public:
    void StateActuatorFirst_Enter() { TrackCall(__FUNCTION__); }
    void StateActuatorFirst_Exit() { TrackCall(__FUNCTION__); }
    void StateActuatorSecond_Enter() { TrackCall(__FUNCTION__); }
    void StateActuatorSecond_Exit() { TrackCall(__FUNCTION__); }
    void StateActuatorThird_Enter() { TrackCall(__FUNCTION__); }
    void StateActuatorThird_Exit() { TrackCall(__FUNCTION__); }
    void StateActuatorThirdEmbed_Enter() { TrackCall(__FUNCTION__); }
    void StateActuatorThirdEmbed_Exit() { TrackCall(__FUNCTION__); }
    void StateActuatorThirdEmbed_First_Enter() { TrackCall(__FUNCTION__); }
    void StateActuatorThirdEmbed_First_Exit() { TrackCall(__FUNCTION__); }
    void StateActuatorThirdEmbed_FirstEmbed_Enter() { TrackCall(__FUNCTION__); }
    void StateActuatorThirdEmbed_FirstEmbed_Exit() { TrackCall(__FUNCTION__); }
    void StateActuatorThirdEmbed_Second_Enter() { TrackCall(__FUNCTION__); }
    void StateActuatorThirdEmbed_Second_Exit() { TrackCall(__FUNCTION__); }

    const std::vector<std::string>& GetCalls() const { return Calls; }

private:
    void TrackCall(const std::string& function_name)
    {
        auto extracted = test::ExtractFunctionName(function_name);

        std::cout << extracted << std::endl;
        Calls.push_back(extracted);
    }

    std::vector<std::string> Calls;
};

TEST(GochartLevelParallel, Test)
{
    using namespace gochart::GochartLevelParallel;

    Delegate delegate;
    auto gochart = gochart::GochartLevelParallel::Create(&delegate);

    gochart->Activate();

    test::CompareStates<EState>(std::array{EState::__Root__,
                                           EState::ActuatorFirst,
                                           EState::ActuatorSecond,
                                           EState::ActuatorThirdEmbed,
                                           EState::ActuatorThirdEmbed_FirstEmbed,
                                           EState::ActuatorThirdEmbed_Second},
                                gochart->GetCurrentStates());

    gochart->Deactivate();

    // Now we compare the calls.
    test::CompareCalls(
        std::array{
            std::string("StateActuatorFirst_Enter"),
            std::string("StateActuatorSecond_Enter"),
            std::string("StateActuatorThird_Enter"),
            std::string("StateActuatorThirdEmbed_Enter"),
            std::string("StateActuatorThirdEmbed_First_Enter"),
            std::string("StateActuatorThirdEmbed_FirstEmbed_Enter"),
            std::string("StateActuatorThirdEmbed_Second_Enter"),

            std::string("StateActuatorThirdEmbed_Second_Exit"),
            std::string("StateActuatorThirdEmbed_FirstEmbed_Exit"),
            std::string("StateActuatorThirdEmbed_First_Exit"),
            std::string("StateActuatorThirdEmbed_Exit"),
            std::string("StateActuatorThird_Exit"),
            std::string("StateActuatorSecond_Exit"),
            std::string("StateActuatorFirst_Exit"),
        },
        delegate.GetCalls());
}
