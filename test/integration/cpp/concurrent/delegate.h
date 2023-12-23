#include "concurrent.forward_declare.h"

#include <functional>
#include <iostream>
#include <map>
#include <memory>
#include <vector>

namespace
{

std::string ExtractFunctionName(const std::string& function_name)
{
    size_t pos = function_name.rfind("::");
    if (pos != std::string::npos)
    {
        // +2 to skip "::".
        return function_name.substr(pos + 2);
    }
    // Return the original name if "::" not found.
    return function_name;
}

} // namespace

namespace test
{

enum class EActuatorType
{
    Invalid,
    Spawn,
    Aiming,
};

class Delegate
{
public:
    using Callback = std::function<void()>;

    Delegate();
    ~Delegate();

    gochart::Concurrent::Gochart<Delegate>& GetGochart();
    const gochart::Concurrent::Gochart<Delegate>& GetGochart() const;
    const std::vector<std::string>& GetCalls() const { return Calls; }

    void SetCallback(const std::string& key, Callback&& callback)
    {
        Callbacks[key] = std::move(callback);
    }

    void Begin();
    void End();

    // Gochart delegates begin.
    void StateRoot_Exit() { TrackCall(__FUNCTION__); }

    void StateMainActuator_Exit() { TrackCall(__FUNCTION__); }
    void StateReady_Enter() { TrackCall(__FUNCTION__); }
    void StateAttack_Enter() { TrackCall(__FUNCTION__); }
    void StateAttack_Exit() { TrackCall(__FUNCTION__); }

    void StateSpawnActuator_Exit() { TrackCall(__FUNCTION__); }
    void StateSpawnActive_Exit() { TrackCall(__FUNCTION__); }
    void StateSpawnDone_Enter() { TrackCall(__FUNCTION__); }

    void StateAimingActuator_Enter() { TrackCall(__FUNCTION__); }
    void StateAimingActuator_Exit() { TrackCall(__FUNCTION__); }
    void StateAimingActive_Enter() { TrackCall(__FUNCTION__); }
    void StateAimingActive_Exit() { TrackCall(__FUNCTION__); }
    void StateAimingSearcher_CaptureUpdate(float dt) { TrackCall(__FUNCTION__); }
    void StateAimingAnalyzer_Enter() { TrackCall(__FUNCTION__); }
    void StateAimingAnalyzer_Exit() { TrackCall(__FUNCTION__); }
    void StateAimingAnalyzer_CaptureTargetAcquired() { TrackCall(__FUNCTION__); }
    void StateAimingDispatcher_Exit() { TrackCall(__FUNCTION__); }

    // Gochart delegates end.

private:
    void TrackCall(const std::string& function_name)
    {
        auto extracted = ExtractFunctionName(function_name);

        std::cout << extracted << std::endl;
        Calls.push_back(extracted);

        // See if we need to call the callback.
        if (Callbacks.contains(extracted))
        {
            Callbacks[extracted]();
        }
    }

    std::unique_ptr<gochart::Concurrent::Gochart<Delegate>> Gochart;
    std::vector<std::string> Calls;
    std::map<std::string, Callback> Callbacks;
};

} // namespace test
