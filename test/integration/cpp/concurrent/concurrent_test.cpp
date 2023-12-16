#include "delegate.h"

#include "concurrent.define.h"

#include <gtest/gtest.h>

#include <array>
#include <span>
#include <sstream>

namespace
{

using namespace gochart::Concurrent;

std::string ToString(const std::span<const EState> states)
{
    std::stringstream ss;
    ss << "[";
    for (size_t i = 0; i < states.size(); i++)
    {
        if (i > 0)
        {
            ss << ", ";
        }

        EState state = states[i];
        ss << ToString(state);
    }
    ss << "]";

    return ss.str();
}

std::string PrintContext(const std::span<const EState> want, const std::span<const EState> got)
{
    std::stringstream ss;
    ss << std::endl << "Want: " << ToString(want) << std::endl << "Got: " << ToString(got);

    return ss.str();
}

void CompareStates(const std::span<const EState> want, const std::span<const EState> got)
{
    using namespace gochart::Concurrent;

    ASSERT_EQ(want.size(), got.size()) << PrintContext(want, got);

    for (size_t i = 0; i < want.size(); i++)
    {
        ASSERT_EQ(want[i], got[i]) << "Want: " << ToString(want[i]) << ", Got: " << ToString(got[i])
                                   << PrintContext(want, got);
    }
}

} // namespace

namespace test
{

Delegate::Delegate() {}

Delegate::~Delegate() = default;

void Delegate::Begin()
{
    Gochart = gochart::Concurrent::Create(this);
    Gochart->Activate();
}

void Delegate::End() { Gochart->Deactivate(); }

gochart::Concurrent::Gochart<Delegate>& Delegate::GetGochart()
{
    assert(Gochart);
    return *Gochart;
}

const gochart::Concurrent::Gochart<Delegate>& Delegate::GetGochart() const
{
    assert(Gochart);
    return *Gochart;
}

} // namespace test

TEST(ConcurrentTest, Hello)
{
    using namespace gochart::Concurrent;

    test::Delegate delegate;
    delegate.Begin();

    auto& gochart = delegate.GetGochart();

    CompareStates(std::array{EState::Root, EState::Idle, EState::SpawnInactive},
                  gochart.GetCurrentStates());
    EXPECT_TRUE(gochart.IsInStateExact(EState::Root));
    EXPECT_TRUE(gochart.IsInStateExact(EState::Idle));
    EXPECT_TRUE(gochart.IsInStateExact(EState::SpawnInactive));

    gochart.TriggerActivateActuator(test::EActuatorType::Spawn, true);
    CompareStates(std::array{EState::Root, EState::Idle, EState::SpawnIdle},
                  gochart.GetCurrentStates());
    EXPECT_TRUE(gochart.IsInStateExact(EState::Root));
    EXPECT_TRUE(gochart.IsInStateExact(EState::Idle));
    EXPECT_TRUE(gochart.IsInStateExact(EState::SpawnIdle));

    gochart.TriggerStartSpawning();
    CompareStates(std::array{EState::Root, EState::WaitingOnSpawning, EState::SpawnInProgress},
                  gochart.GetCurrentStates());
    EXPECT_TRUE(gochart.IsInStateExact(EState::Root));
    EXPECT_TRUE(gochart.IsInStateExact(EState::WaitingOnSpawning));
    EXPECT_TRUE(gochart.IsInStateExact(EState::SpawnInProgress));

    // TODO(cdc): Activate when null transitions are done.
    /* gochart.TriggerSpawned(); */
    /* CompareStates(std::array{EState::Root, EState::Ready, EState::SpawnIdle}, */
    /*               gochart.GetCurrentStates()); */
    /* EXPECT_TRUE(gochart.IsInStateExact(EState::Root)); */
    /* EXPECT_TRUE(gochart.IsInStateExact(EState::Ready)); */
    /* EXPECT_TRUE(gochart.IsInStateExact(EState::SpawnIdle)); */

    delegate.End();
}
