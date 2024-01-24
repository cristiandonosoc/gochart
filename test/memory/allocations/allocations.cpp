#include <new>
#include <cstdint>
#include <cstdlib>
#include <iostream>
#include <format>
#include <string>
#include <vector>
#include <sstream>

// Override the new/delete operators to count the amount of allocations.
// https://stackoverflow.com/questions/9927856/how-to-use-operator-new-to-count-number-of-times-of-dynamic-memory-allocation

int number_of_allocs = 0;

void* operator new(std::size_t size) {
  ++number_of_allocs;
  void *p = malloc(size);
  return p;
}

void* operator new  [](std::size_t size) {
  ++number_of_allocs;
  void *p = malloc(size);
  return p;
}

void* operator new  [](std::size_t size, const std::nothrow_t&) {
  ++number_of_allocs;
  return malloc(size);
}
void* operator new   (std::size_t size, const std::nothrow_t&) {
  ++number_of_allocs;
  return malloc(size);
}


void operator delete(void* ptr) { free(ptr); }
void operator delete (void* ptr, const std::nothrow_t&) { free(ptr); }
void operator delete[](void* ptr) { free(ptr); }
void operator delete[](void* ptr, const std::nothrow_t&) { free(ptr); }

namespace test
{

enum class EActuatorType : uint8_t
{
	Spawn,
	Aiming,
};

} // namespace test

#include "simple.gochart.define.h"
#include "parallel.gochart.define.h"

class Delegate
{
public:
    // Simple delegates begin.
    void StateDoor_CaptureUpdate(bool tick) { }
    void StateDoorInactive_CaptureUpdate(bool tick) { }
    void StateDoorOpened_CaptureUpdate(bool tick) { }
    void StateDoorMoving_EnterToggleDoor(bool open, bool fast) { }
    void StateWindowInitial_Enter() { }
    void StateWindowInitial_Exit() { }
    void StateWindowIdle_Enter() { }

    void ActionFunction(bool active) { }
    void NullTransitionCall() { }
    // Simple delegates end.


    // Parallel delegates begin.
    void StateRoot_Exit() { }

    void StateMainActuator_Exit() { }
    void StateReady_Enter() { }
    void StateAttack_Enter() { }
    void StateAttack_Exit() { }

    void StateSpawnActuator_Exit() { }
    void StateSpawnActive_Exit() { }
    void StateSpawnDone_Enter() { }

    void StateAimingActuator_Enter() { }
    void StateAimingActuator_Exit() { }
    void StateAimingActive_Enter() { }
    void StateAimingActive_Exit() { }
    void StateAimingSearcher_CaptureUpdate(float dt) { }
    void StateAimingAnalyzer_Enter() { }
    void StateAimingAnalyzer_Exit() { }
    void StateAimingAnalyzer_CaptureTargetAcquired() { }
    void StateAimingDispatcher_Exit() { }
	// Parallel delegates end.
};

int main() {
	int alloc_start = 0;
	int alloc_before_simple = 0;
	int alloc_after_simple = 0;
	int alloc_before_parallel = 0;
	int alloc_after_parallel = 0;
	int alloc_after = 0;

	// Make sure we have some allocations to compare.
	std::vector<std::string> vector;
	for (int i = 0; i < 1024; i++)
	{
		vector.push_back(std::format("{}", i));
	}

	alloc_start = number_of_allocs;

	{
		Delegate delegate;
		gochart::Simple::Gochart simple_gochart(&delegate);

		alloc_before_simple = number_of_allocs;

    	simple_gochart.Activate();
    	simple_gochart.TriggerUpdate(true);
    	simple_gochart.TriggerActivateDoorActuator(false);
    	simple_gochart.TriggerUpdate(true);
    	simple_gochart.TriggerActivateDoorActuator(true);
    	simple_gochart.TriggerUpdate(true);
    	simple_gochart.TriggerUpdate(false);
    	simple_gochart.TriggerToggleDoor(true, false);
    	simple_gochart.TriggerUpdate(true);
    	simple_gochart.TriggerToggleDoor(true, false);
		simple_gochart.Deactivate();

		alloc_after_simple = number_of_allocs;
	}

	{
		Delegate delegate;
		gochart::Parallel::Gochart parallel_gochart(&delegate);

		alloc_before_parallel = number_of_allocs;

		parallel_gochart.Activate();
    	parallel_gochart.TriggerActivateActuator(test::EActuatorType::Spawn, true);
    	parallel_gochart.TriggerStartSpawning();
    	parallel_gochart.TriggerActivateActuator(test::EActuatorType::Aiming, true);
    	parallel_gochart.TriggerUpdate(0.16f);
    	parallel_gochart.TriggerUpdate(0.16f);
    	parallel_gochart.TriggerUpdate(0.16f);
    	parallel_gochart.TriggerTargetDispatched();
		parallel_gochart.Deactivate();

		alloc_after_parallel = number_of_allocs;
	}



	alloc_after = number_of_allocs;


	std::cout << "alloc_start: " << alloc_start << std::endl;
	std::cout << "alloc_before_simple: " << alloc_before_simple << std::endl;
	std::cout << "alloc_after_simple: " << alloc_after_simple << std::endl;
	std::cout << "alloc_before_parallel: " << alloc_before_parallel << std::endl;
	std::cout << "alloc_after_parallel: " << alloc_after_parallel << std::endl;
	std::cout << "alloc_after: " << alloc_after << std::endl;
}
