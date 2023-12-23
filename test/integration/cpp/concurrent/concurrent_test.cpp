#include "delegate.h"

#include "concurrent.define.h"

#include <gtest/gtest.h>

#include <array>
#include <source_location>
#include <span>
#include <sstream>

namespace
{

using namespace gochart::Concurrent;

template <typename T>
std::string ToString(const T& t)
{
    std::stringstream ss;
    ss << t;
    return ss.str();
}

std::string ToString(const std::source_location& location)
{
    std::stringstream ss;
    ss << location.file_name() << ":" << location.line();
    return ss.str();
}

template <typename T>
std::string PrintSpan(std::span<T> things)
{
    std::stringstream ss;
    ss << "[" << std::endl;
    for (size_t i = 0; i < things.size(); i++)
    {
        const T& thing = things[i];
        ss << "\t" << ToString(thing) << std::endl;
    }
    ss << "]";

    return ss.str();
}

template <typename T>
std::string PrintContext(std::span<T> want, std::span<T> got, const std::source_location& location)
{
    std::stringstream ss;
    ss << std::endl
       << "Want: " << PrintSpan(want) << std::endl
       << "Got: " << PrintSpan(got) << std::endl
       << "Location: " << ToString(location);

    return ss.str();
}

void CompareStates(std::span<const EState> want,
                   std::span<const EState> got,
                   std::source_location location = std::source_location::current())
{
    using namespace gochart::Concurrent;

    ASSERT_EQ(want.size(), got.size()) << PrintContext(want, got, location);

    for (size_t i = 0; i < want.size(); i++)
    {
        ASSERT_EQ(want[i], got[i]) << "Want: " << ToString(want[i]) << ", Got: " << ToString(got[i])
                                   << PrintContext(want, got, location);
    }
}

void CompareCalls(std::span<const std::string> want,
                  std::span<const std::string> got,
                  std::source_location location = std::source_location::current())
{
    ASSERT_EQ(want.size(), got.size()) << PrintContext(want, got, location);

    for (size_t i = 0; i < want.size(); i++)
    {
        ASSERT_EQ(want[i], got[i]) << "Want: " << ToString(want[i]) << ", Got: " << ToString(got[i])
                                   << PrintContext(want, got, location);
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

TEST(ConcurrentTest, SimpleCase)
{
    using namespace gochart::Concurrent;

    test::Delegate delegate;
    delegate.Begin();

    auto& gochart = delegate.GetGochart();

    CompareStates(std::array{EState::Root,
                             EState::Idle,
                             EState::SpawnInactive,
                             EState::AimingInactive,
                             EState::__Invalid__,
                             EState::__Invalid__,
                             EState::__Invalid__},
                  gochart.GetCurrentStates());
    EXPECT_TRUE(gochart.IsInStateExact(EState::Root));
    EXPECT_TRUE(gochart.IsInStateExact(EState::Idle));
    EXPECT_TRUE(gochart.IsInStateExact(EState::SpawnInactive));
    EXPECT_TRUE(gochart.IsInStateExact(EState::AimingInactive));

    gochart.TriggerActivateActuator(test::EActuatorType::Spawn, true);
    CompareStates(std::array{EState::Root,
                             EState::Idle,
                             EState::SpawnIdle,
                             EState::AimingInactive,
                             EState::__Invalid__,
                             EState::__Invalid__,
                             EState::__Invalid__},
                  gochart.GetCurrentStates());
    EXPECT_TRUE(gochart.IsInStateExact(EState::Root));
    EXPECT_TRUE(gochart.IsInStateExact(EState::Idle));
    EXPECT_TRUE(gochart.IsInStateExact(EState::SpawnIdle));
    EXPECT_TRUE(gochart.IsInStateExact(EState::AimingInactive));

    gochart.TriggerStartSpawning();
    CompareStates(std::array{EState::Root,
                             EState::WaitingOnSpawning,
                             EState::SpawnInProgress,
                             EState::AimingInactive,
                             EState::__Invalid__,
                             EState::__Invalid__,
                             EState::__Invalid__},
                  gochart.GetCurrentStates());
    EXPECT_TRUE(gochart.IsInStateExact(EState::Root));
    EXPECT_TRUE(gochart.IsInStateExact(EState::WaitingOnSpawning));
    EXPECT_TRUE(gochart.IsInStateExact(EState::SpawnInProgress));
    EXPECT_TRUE(gochart.IsInStateExact(EState::AimingInactive));

    gochart.TriggerSpawned();
    CompareStates(std::array{EState::Root,
                             EState::Ready,
                             EState::SpawnIdle,
                             EState::AimingInactive,
                             EState::__Invalid__,
                             EState::__Invalid__,
                             EState::__Invalid__},
                  gochart.GetCurrentStates());
    EXPECT_TRUE(gochart.IsInStateExact(EState::Root));
    EXPECT_TRUE(gochart.IsInStateExact(EState::Ready));
    EXPECT_TRUE(gochart.IsInStateExact(EState::SpawnIdle));
    EXPECT_TRUE(gochart.IsInStateExact(EState::AimingInactive));

    delegate.End();

    CompareCalls(
        std::array{
            std::string("StateAimingActuator_Enter"),
            std::string("StateSpawnDone_Enter"),
            std::string("StateReady_Enter"),

            std::string("StateMainActuator_Exit"),
            std::string("StateSpawnActive_Exit"),
            std::string("StateSpawnActuator_Exit"),
            std::string("StateAimingActuator_Exit"),
            std::string("StateRoot_Exit"),
        },
        delegate.GetCalls());
}

TEST(ConcurrentTest, ComplexCase)
{
    using namespace gochart::Concurrent;

    test::Delegate delegate;
    delegate.Begin();

    auto& gochart = delegate.GetGochart();
    EXPECT_EQ(delegate.GetCalls().size(), 1);
    CompareStates(std::array{EState::Root,
                             EState::Idle,
                             EState::SpawnInactive,
                             EState::AimingInactive,
                             EState::__Invalid__,
                             EState::__Invalid__,
                             EState::__Invalid__},
                  gochart.GetCurrentStates());

    gochart.TriggerActivateActuator(test::EActuatorType::Spawn, true);
    EXPECT_EQ(delegate.GetCalls().size(), 1);
    CompareStates(std::array{EState::Root,
                             EState::Idle,
                             EState::SpawnIdle,
                             EState::AimingInactive,
                             EState::__Invalid__,
                             EState::__Invalid__,
                             EState::__Invalid__},
                  gochart.GetCurrentStates());

    gochart.TriggerStartSpawning();
    EXPECT_EQ(delegate.GetCalls().size(), 1);
    CompareStates(std::array{EState::Root,
                             EState::WaitingOnSpawning,
                             EState::SpawnInProgress,
                             EState::AimingInactive,
                             EState::__Invalid__,
                             EState::__Invalid__,
                             EState::__Invalid__},
                  gochart.GetCurrentStates());

    gochart.TriggerActivateActuator(test::EActuatorType::Aiming, true);
    EXPECT_EQ(delegate.GetCalls().size(), 3);
    CompareStates(std::array{EState::Root,
                             EState::WaitingOnSpawning,
                             EState::SpawnInProgress,
                             EState::AimingActive,
                             EState::AimingSearcher,
                             EState::AimingAnalyzer,
                             EState::AimingDispatcherIdle},
                  gochart.GetCurrentStates());

    gochart.TriggerUpdate(0.16f);
    EXPECT_EQ(delegate.GetCalls().size(), 4);
    CompareStates(std::array{EState::Root,
                             EState::WaitingOnSpawning,
                             EState::SpawnInProgress,
                             EState::AimingActive,
                             EState::AimingSearcher,
                             EState::AimingAnalyzer,
                             EState::AimingDispatcherIdle},
                  gochart.GetCurrentStates());

    delegate.SetCallback("StateAimingSearcher_CaptureUpdate",
                         [&gochart]() { gochart.TriggerTargetAcquired(); });
    gochart.TriggerUpdate(0.16f);
    EXPECT_EQ(delegate.GetCalls().size(), 6);
    CompareStates(std::array{EState::Root,
                             EState::WaitingOnSpawning,
                             EState::SpawnInProgress,
                             EState::AimingActive,
                             EState::AimingSearcher,
                             EState::AimingAnalyzer,
                             EState::AimingDispatcherIdle},
                  gochart.GetCurrentStates());

    delegate.SetCallback("StateAimingAnalyzer_CaptureTargetAcquired",
                         [&gochart]() { gochart.TriggerTargetConfirmed(); });
    gochart.TriggerUpdate(0.16f);
    EXPECT_EQ(delegate.GetCalls().size(), 8);
    CompareStates(std::array{EState::Root,
                             EState::WaitingOnSpawning,
                             EState::SpawnInProgress,
                             EState::AimingActive,
                             EState::AimingSearcher,
                             EState::AimingAnalyzer,
                             EState::AimingDispatcherDispatching},
                  gochart.GetCurrentStates());

    gochart.TriggerTargetDispatched();
    EXPECT_EQ(delegate.GetCalls().size(), 9);
    CompareStates(std::array{EState::Root,
                             EState::Attack,
                             EState::SpawnInProgress,
                             EState::AimingActive,
                             EState::AimingSearcher,
                             EState::AimingAnalyzer,
                             EState::AimingDispatcherIdle},
                  gochart.GetCurrentStates());

    delegate.End();

    // Now we compare the calls.
    CompareCalls(
        std::array{
            std::string("StateAimingActuator_Enter"),
            std::string("StateAimingActive_Enter"),
            std::string("StateAimingAnalyzer_Enter"),
            std::string("StateAimingSearcher_CaptureUpdate"),
            std::string("StateAimingSearcher_CaptureUpdate"),
            std::string("StateAimingAnalyzer_CaptureTargetAcquired"),
            std::string("StateAimingSearcher_CaptureUpdate"),
            std::string("StateAimingAnalyzer_CaptureTargetAcquired"),
            std::string("StateAttack_Enter"),

            std::string("StateAttack_Exit"),
            std::string("StateMainActuator_Exit"),
            std::string("StateSpawnActive_Exit"),
            std::string("StateSpawnActuator_Exit"),
            std::string("StateAimingAnalyzer_Exit"),
            std::string("StateAimingDispatcher_Exit"),
            std::string("StateAimingActive_Exit"),
            std::string("StateAimingActuator_Exit"),
            std::string("StateRoot_Exit"),
        },
        delegate.GetCalls());
}
