#include "delegate.h"

#include "concurrent.define.h"

#include <gtest/gtest.h>

#include <array>
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

template <typename T>
std::string PrintSpan(const std::span<T> things)
{
    std::stringstream ss;
    ss << "[";
    for (size_t i = 0; i < things.size(); i++)
    {
        if (i > 0)
        {
            ss << ", ";
        }

        const T& thing = things[i];
        ss << ToString(thing);
    }
    ss << "]";

    return ss.str();
}

template <typename T>
std::string PrintContext(const std::span<T> want, const std::span<T> got)
{
    std::stringstream ss;
    ss << std::endl << "Want: " << PrintSpan(want) << std::endl << "Got: " << PrintSpan(got);

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

void CompareStrings(const std::span<const std::string> want, const std::span<const std::string> got)
{
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

TEST(ConcurrentTest, SimpleCase)
{
    using namespace gochart::Concurrent;

    test::Delegate delegate;
    delegate.Begin();

    auto& gochart = delegate.GetGochart();

    CompareStates(std::array{EState::Root, EState::Idle, EState::SpawnInactive,
                             EState::AimingInactive, EState::Invalid, EState::Invalid,
                             EState::Invalid},
                  gochart.GetCurrentStates());
    EXPECT_TRUE(gochart.IsInStateExact(EState::Root));
    EXPECT_TRUE(gochart.IsInStateExact(EState::Idle));
    EXPECT_TRUE(gochart.IsInStateExact(EState::SpawnInactive));
    EXPECT_TRUE(gochart.IsInStateExact(EState::AimingInactive));

    gochart.TriggerActivateActuator(test::EActuatorType::Spawn, true);
    CompareStates(std::array{EState::Root, EState::Idle, EState::SpawnIdle, EState::AimingInactive,
                             EState::Invalid, EState::Invalid, EState::Invalid},
                  gochart.GetCurrentStates());
    EXPECT_TRUE(gochart.IsInStateExact(EState::Root));
    EXPECT_TRUE(gochart.IsInStateExact(EState::Idle));
    EXPECT_TRUE(gochart.IsInStateExact(EState::SpawnIdle));
    EXPECT_TRUE(gochart.IsInStateExact(EState::AimingInactive));

    gochart.TriggerStartSpawning();
    CompareStates(std::array{EState::Root, EState::WaitingOnSpawning, EState::SpawnInProgress,
                             EState::AimingInactive, EState::Invalid, EState::Invalid,
                             EState::Invalid},
                  gochart.GetCurrentStates());
    EXPECT_TRUE(gochart.IsInStateExact(EState::Root));
    EXPECT_TRUE(gochart.IsInStateExact(EState::WaitingOnSpawning));
    EXPECT_TRUE(gochart.IsInStateExact(EState::SpawnInProgress));
    EXPECT_TRUE(gochart.IsInStateExact(EState::AimingInactive));

    gochart.TriggerSpawned();
    CompareStates(std::array{EState::Root, EState::Ready, EState::SpawnIdle, EState::AimingInactive,
                             EState::Invalid, EState::Invalid, EState::Invalid},
                  gochart.GetCurrentStates());
    EXPECT_TRUE(gochart.IsInStateExact(EState::Root));
    EXPECT_TRUE(gochart.IsInStateExact(EState::Ready));
    EXPECT_TRUE(gochart.IsInStateExact(EState::SpawnIdle));
    EXPECT_TRUE(gochart.IsInStateExact(EState::AimingInactive));

    delegate.End();

    CompareStrings(
        std::array{
            std::string("StateSpawnDone_Enter"),
            std::string("StateReady_Enter"),
        },
        delegate.GetCalls());
}

/* TEST(ConcurrentTest, ComplexCase) */
/* { */
/*     using namespace gochart::Concurrent; */

/*     test::Delegate delegate; */
/*     delegate.Begin(); */

/*     auto& gochart = delegate.GetGochart(); */

/*     CompareStates(std::array{EState::Root, EState::Idle, EState::SpawnInactive, */
/*                              EState::AimingInactive, EState::Invalid, EState::Invalid, */
/*                              EState::Invalid}, */
/*                   gochart.GetCurrentStates()); */
/*     EXPECT_TRUE(gochart.IsInStateExact(EState::Root)); */
/*     EXPECT_TRUE(gochart.IsInStateExact(EState::Idle)); */
/*     EXPECT_TRUE(gochart.IsInStateExact(EState::SpawnInactive)); */
/*     EXPECT_TRUE(gochart.IsInStateExact(EState::AimingInactive)); */

/*     gochart.TriggerActivateActuator(test::EActuatorType::Spawn, true); */
/*     CompareStates(std::array{EState::Root, EState::Idle, EState::SpawnIdle, EState::AimingInactive, */
/*                              EState::Invalid, EState::Invalid, EState::Invalid}, */
/*                   gochart.GetCurrentStates()); */
/*     EXPECT_TRUE(gochart.IsInStateExact(EState::Root)); */
/*     EXPECT_TRUE(gochart.IsInStateExact(EState::Idle)); */
/*     EXPECT_TRUE(gochart.IsInStateExact(EState::SpawnIdle)); */
/*     EXPECT_TRUE(gochart.IsInStateExact(EState::AimingInactive)); */

/*     gochart.TriggerStartSpawning(); */
/*     CompareStates(std::array{EState::Root, EState::WaitingOnSpawning, EState::SpawnInProgress, */
/*                              EState::AimingInactive, EState::Invalid, EState::Invalid, */
/*                              EState::Invalid}, */
/*                   gochart.GetCurrentStates()); */
/*     EXPECT_TRUE(gochart.IsInStateExact(EState::Root)); */
/*     EXPECT_TRUE(gochart.IsInStateExact(EState::WaitingOnSpawning)); */
/*     EXPECT_TRUE(gochart.IsInStateExact(EState::SpawnInProgress)); */
/*     EXPECT_TRUE(gochart.IsInStateExact(EState::AimingInactive)); */

/*     gochart.TriggerActivateActuator(test::EActuatorType::Aiming, true); */
/*     CompareStates(std::array{EState::Root, EState::WaitingOnSpawning, EState::SpawnInProgress, */
/*                              EState::AimingActive, EState::AimingSearcher, EState::AimingAnalyzer, */
/*                              EState::AimingDispatcherIdle}, */
/*                   gochart.GetCurrentStates()); */
/*     EXPECT_TRUE(gochart.IsInStateExact(EState::Root)); */
/*     EXPECT_TRUE(gochart.IsInStateExact(EState::WaitingOnSpawning)); */
/*     EXPECT_TRUE(gochart.IsInStateExact(EState::SpawnInProgress)); */
/*     EXPECT_TRUE(gochart.IsInStateExact(EState::AimingActive)); */

/*     delegate.End(); */
/* } */
