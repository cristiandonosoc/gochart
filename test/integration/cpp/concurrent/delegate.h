#include "concurrent.forward_declare.h"

#include <iostream>
#include <memory>

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

private:
    std::unique_ptr<gochart::Concurrent::Gochart<Delegate>> Gochart;
};
