#include "concurrent.forward_declare.h"

#include <iostream>
#include <memory>

namespace test
{

enum class EActuatorType
{
    Invalid,
    Spawn,
};

class Delegate
{
public:
    Delegate();
    ~Delegate();

    void Begin();
    void End();

    // Gochart delegates begin.
    void StateReady_Enter() { std::cout << __FUNCTION__ << std::endl; }
    void StateSpawnDone_Enter() { std::cout << __FUNCTION__ << std::endl; }
    // Gochart delegates end.


    gochart::Concurrent::Gochart<Delegate>& GetGochart();
    const gochart::Concurrent::Gochart<Delegate>& GetGochart() const;

private:
    std::unique_ptr<gochart::Concurrent::Gochart<Delegate>> Gochart;
};

} // namespace test
