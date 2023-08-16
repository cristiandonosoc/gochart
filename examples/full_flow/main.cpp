#include "statechart.generated.h"

#include <iostream>

class Delegate
{
public:
    void StateStateA_OnEnter() { std::cout << __FUNCTION__ << std::endl; }
    void StateStateA_OnEnterByTrigger1(int foo, float bar)
    {
        std::cout << __FUNCTION__ << ", foo: " << foo << ", bar: " << bar << std::endl;
    }
    void StateStateA_OnExit() { std::cout << __FUNCTION__ << std::endl; }
    void StateStateA_OnExitByTrigger1(int foo, float bar)
    {
        std::cout << __FUNCTION__ << ", foo: " << foo << ", bar: " << bar << std::endl;
    }
    void StateStateB_OnExitByTrigger1(int foo, float bar)
    {
        std::cout << __FUNCTION__ << ", foo: " << foo << ", bar: " << bar << std::endl;
    }
    void StateStateC_OnEnter() { std::cout << __FUNCTION__ << std::endl; }
    void StateStateC_OnEnterByTrigger1(int foo, float bar)
    {
        std::cout << __FUNCTION__ << ", foo: " << foo << ", bar: " << bar << std::endl;
    }
};

int main()
{
    Delegate delegate;
    auto statechart = gochart::StatechartSimple<Delegate>::Create(&delegate);

    statechart->Activate();

    std::cout << "Current State: "
              << gochart::StatechartSimpleImpl::ToString(statechart->GetCurrentState())
              << std::endl;

    statechart->TriggerTrigger1(2, 2.5f);

    std::cout << "Current State: "
              << gochart::StatechartSimpleImpl::ToString(statechart->GetCurrentState())
              << std::endl;

    statechart->Deactivate();
}
