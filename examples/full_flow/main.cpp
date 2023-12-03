#include "statechart.generated.h"

#include <iostream>

class Delegate
{
public:
    void StateStateA_Enter() { std::cout << __FUNCTION__ << std::endl; }
    void StateStateA_EnterTrigger1(int foo, float bar)
    {
        std::cout << __FUNCTION__ << ", foo: " << foo << ", bar: " << bar << std::endl;
    }
    void StateStateA_Exit() { std::cout << __FUNCTION__ << std::endl; }
    void StateStateA_ExitTrigger1(int foo, float bar)
    {
        std::cout << __FUNCTION__ << ", foo: " << foo << ", bar: " << bar << std::endl;
    }
    void StateStateA_CaptureTrigger1(int foo, float bar)
    {
        std::cout << __FUNCTION__ << ", foo: " << foo << ", bar: " << bar << std::endl;
    }

    void StateStateB_ExitTrigger1(int foo, float bar)
    {
        std::cout << __FUNCTION__ << ", foo: " << foo << ", bar: " << bar << std::endl;
    }
    void StateStateC_Enter() { std::cout << __FUNCTION__ << std::endl; }
    void StateStateC_EnterTrigger1(int foo, float bar)
    {
        std::cout << __FUNCTION__ << ", foo: " << foo << ", bar: " << bar << std::endl;
    }
};

int main()
{
    Delegate delegate;
    auto statechart = gochart::StatechartSimple<Delegate>::Create(&delegate);

    statechart->Activate();

    //std::cout << "Current State: "
    //          << gochart::ToString(statechart->GetCurrentState())
    //          << std::endl;

    statechart->TriggerTrigger1(2, 2.5f);

    //std::cout << "Current State: "
    //          << gochart::ToString(statechart->GetCurrentState())
    //          << std::endl;

    statechart->Deactivate();
}
