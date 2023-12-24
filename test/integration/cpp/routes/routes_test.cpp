#include "test/test_support.h"

#include <gtest/gtest.h>

#include <vector>

#include <routes.define.h>

class Delegate
{
public:
    void StateFirst_Enter() { TrackCall(__FUNCTION__); }
    void StateFirst_Exit() { TrackCall(__FUNCTION__); }
    void StateFirstSon1_Enter() { TrackCall(__FUNCTION__); }
    void StateFirstSon1_Exit() { TrackCall(__FUNCTION__); }
    void StateFirstSon2_Enter() { TrackCall(__FUNCTION__); }
    void StateFirstSon2_Exit() { TrackCall(__FUNCTION__); }
    void StateSecond_Enter() { TrackCall(__FUNCTION__); }
    void StateSecond_Exit() { TrackCall(__FUNCTION__); }
    void StateSecondSon1_Enter() { TrackCall(__FUNCTION__); }
    void StateSecondSon1_Exit() { TrackCall(__FUNCTION__); }
    void StateThird_Enter() { TrackCall(__FUNCTION__); }
    void StateThird_Exit() { TrackCall(__FUNCTION__); }

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

TEST(Routes, Test)
{
    using namespace gochart::Routes;

    Delegate delegate;
    auto gochart = gochart::Routes::Create(&delegate);

    gochart->Activate();
    EXPECT_TRUE(gochart->IsInState(EState::FirstSon2))
        << "CURRENT STATES: " << test::PrintSpan<EState>(gochart->GetCurrentStates());

    gochart->TriggerGoToSecond();
    EXPECT_TRUE(gochart->IsInState(EState::SecondSon1))
        << "CURRENT STATES: " << test::PrintSpan<EState>(gochart->GetCurrentStates());

    gochart->TriggerGoToFirst();
    EXPECT_TRUE(gochart->IsInState(EState::FirstSon1))
        << "CURRENT STATES: " << test::PrintSpan<EState>(gochart->GetCurrentStates());

    gochart->TriggerGoToThird();
    EXPECT_TRUE(gochart->IsInState(EState::Third))
        << "CURRENT STATES: " << test::PrintSpan<EState>(gochart->GetCurrentStates());

    gochart->TriggerGoToFirst();
    EXPECT_TRUE(gochart->IsInState(EState::FirstSon2))
        << "CURRENT STATES: " << test::PrintSpan<EState>(gochart->GetCurrentStates());

    gochart->Deactivate();
}
