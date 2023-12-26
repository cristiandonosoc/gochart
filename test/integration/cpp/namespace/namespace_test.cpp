#include "test/test_support.h"

#include <gtest/gtest.h>

#include <vector>

#include <original_name.define.h>

class Delegate
{
public:
    void StateFoo_Enter() { TrackCall(__FUNCTION__); }
    void StateFoo_Exit() { TrackCall(__FUNCTION__); }

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

TEST(Namespace, Test)
{
    using namespace gochart::ChangedName;

    Delegate delegate;
    auto gochart = gochart::ChangedName::Create(&delegate);

    gochart->Activate();
    EXPECT_TRUE(gochart->IsInState(EState::Foo))
        << "CURRENT STATES: " << test::PrintSpan<EState>(gochart->GetCurrentStates());

    gochart->Deactivate();

    test::CompareCalls(
        std::array{
            std::string("StateFoo_Enter"),
            std::string("StateFoo_Exit"),
        },
        delegate.GetCalls());
}
