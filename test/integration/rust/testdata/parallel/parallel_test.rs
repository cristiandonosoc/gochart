#[path = "parallel.gochart.rs"]
mod parallel;

use parallel::{Delegate, Leaf, StateId, REGION_COUNT};

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Actuator {
    Spawn,
    Aiming,
}

#[derive(Default)]
struct Robot {
    chart: parallel::Chart,
    calls: Vec<String>,
    spawning_done: bool,
    updates: Vec<&'static str>,
}

impl Robot {
    fn take_calls(&mut self) -> Vec<String> {
        std::mem::take(&mut self.calls)
    }
}

macro_rules! record {
    ($($name:ident),* $(,)?) => {
        $(fn $name(&mut self) { self.calls.push(stringify!($name).to_string()); })*
    };
}

impl Delegate for Robot {
    fn chart(&mut self) -> &mut parallel::Chart {
        &mut self.chart
    }

    record!(
        root_exit, main_actuator_exit, ready_enter, attack_enter, attack_exit,
        spawn_actuator_exit, spawn_active_exit, spawn_done_enter,
        aiming_actuator_enter, aiming_actuator_exit, aiming_active_enter, aiming_active_exit,
        aiming_analyzer_enter, aiming_analyzer_exit, aiming_dispatcher_exit,
    );

    fn root_capture_update(&mut self, _dt: &u32) {
        self.updates.push("root");
    }

    fn aiming_searcher_capture_update(&mut self, _dt: &u32) {
        self.updates.push("searcher");
    }

    fn aiming_analyzer_capture_target_acquired(&mut self) {
        self.calls.push("target_acquired".to_string());
    }

    fn spawn_inactive_spawn_on(&self, actuator: &Actuator, activate: &bool) -> bool {
        *actuator == Actuator::Spawn && *activate
    }

    fn spawn_active_spawn_off(&self, actuator: &Actuator, activate: &bool) -> bool {
        *actuator == Actuator::Spawn && !*activate
    }

    fn aiming_inactive_aiming_on(&self, actuator: &Actuator, activate: &bool) -> bool {
        *actuator == Actuator::Aiming && *activate
    }

    fn aiming_active_aiming_off(&self, actuator: &Actuator, activate: &bool) -> bool {
        *actuator == Actuator::Aiming && !*activate
    }

    fn spawn_done_done(&mut self) {
        // The C++ test raises SpawningActuatorDone here.
        self.spawning_done = true;
    }
}

fn leaves(robot: &Robot) -> [Option<Leaf>; REGION_COUNT] {
    robot.chart.leaves()
}

#[test]
fn regions() {
    let mut robot = Robot::default();
    parallel::activate(&mut robot);

    // Region 0 is the gochart level, which is in the parallel Root rather than a leaf.
    assert_eq!(REGION_COUNT, 7);
    assert_eq!(
        leaves(&robot),
        [None, Some(Leaf::Idle), Some(Leaf::SpawnInactive), Some(Leaf::AimingInactive), None, None, None]
    );
    assert!(robot.chart.is_in(StateId::Root));
    assert!(robot.chart.is_in(StateId::SpawnActuator));
    assert!(!robot.chart.is_in(StateId::AimingActive));
    assert_eq!(robot.take_calls(), vec!["aiming_actuator_enter"]);

    parallel::trigger_activate_actuator(&mut robot, Actuator::Spawn, true);
    assert_eq!(
        leaves(&robot),
        [None, Some(Leaf::Idle), Some(Leaf::SpawnIdle), Some(Leaf::AimingInactive), None, None, None]
    );

    // Both regions handle the trigger.
    parallel::trigger_start_spawning(&mut robot);
    assert_eq!(
        leaves(&robot),
        [None, Some(Leaf::WaitingOnSpawning), Some(Leaf::SpawnInProgress), Some(Leaf::AimingInactive), None, None, None]
    );

    parallel::trigger_spawned(&mut robot);
    assert!(robot.spawning_done);
    assert_eq!(leaves(&robot)[2], Some(Leaf::SpawnIdle));
    assert_eq!(leaves(&robot)[1], Some(Leaf::WaitingOnSpawning));

    // What the C++ chart raises from the null transition.
    parallel::trigger_spawning_actuator_done(&mut robot);
    assert_eq!(leaves(&robot)[1], Some(Leaf::Ready));
    assert_eq!(robot.take_calls(), vec!["spawn_done_enter", "ready_enter"]);

    // Nested parallel state: entering it enters every region.
    parallel::trigger_activate_actuator(&mut robot, Actuator::Aiming, true);
    assert_eq!(
        leaves(&robot),
        [
            None,
            Some(Leaf::Ready),
            Some(Leaf::SpawnIdle),
            None,
            Some(Leaf::AimingSearcher),
            Some(Leaf::AimingAnalyzer),
            Some(Leaf::AimingDispatcherIdle),
        ]
    );
    assert_eq!(robot.take_calls(), vec!["aiming_active_enter", "aiming_analyzer_enter"]);

    // Every region bubbles Update up, but Root is offered it only once.
    parallel::trigger_update(&mut robot, 16);
    assert_eq!(robot.updates, vec!["root", "searcher"]);

    parallel::trigger_target_acquired(&mut robot);
    parallel::trigger_target_confirmed(&mut robot);
    assert_eq!(leaves(&robot)[6], Some(Leaf::AimingDispatcherDispatching));

    // Handled by two regions.
    parallel::trigger_target_dispatched(&mut robot);
    assert_eq!(leaves(&robot)[1], Some(Leaf::Attack));
    assert_eq!(leaves(&robot)[6], Some(Leaf::AimingDispatcherIdle));
    assert_eq!(robot.take_calls(), vec!["target_acquired", "attack_enter"]);

    // Leaving a nested parallel state exits its regions from the bottom up.
    parallel::trigger_activate_actuator(&mut robot, Actuator::Aiming, false);
    assert_eq!(leaves(&robot)[3], Some(Leaf::AimingInactive));
    assert_eq!(leaves(&robot)[4..], [None, None, None]);
    assert_eq!(
        robot.take_calls(),
        vec!["aiming_dispatcher_exit", "aiming_analyzer_exit", "aiming_active_exit"]
    );

    parallel::deactivate(&mut robot);
    assert_eq!(leaves(&robot), [None; REGION_COUNT]);
    assert_eq!(
        robot.take_calls(),
        vec![
            "aiming_actuator_exit",
            "spawn_active_exit",
            "spawn_actuator_exit",
            "attack_exit",
            "main_actuator_exit",
            "root_exit",
        ]
    );
}

#[test]
fn transitions_stay_in_their_region() {
    let mut robot = Robot::default();
    parallel::activate(&mut robot);
    parallel::trigger_activate_actuator(&mut robot, Actuator::Spawn, true);
    parallel::trigger_activate_actuator(&mut robot, Actuator::Aiming, true);
    robot.take_calls();

    // Leaving SpawnActive doesn't touch the other regions.
    parallel::trigger_activate_actuator(&mut robot, Actuator::Spawn, false);
    assert_eq!(
        leaves(&robot),
        [
            None,
            Some(Leaf::Idle),
            Some(Leaf::SpawnInactive),
            None,
            Some(Leaf::AimingSearcher),
            Some(Leaf::AimingAnalyzer),
            Some(Leaf::AimingDispatcherIdle),
        ]
    );
    assert_eq!(robot.take_calls(), vec!["spawn_active_exit"]);
}
