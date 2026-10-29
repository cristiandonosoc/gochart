#[path = "data.gochart.rs"]
mod hauler;

use hauler::{Delegate, Event, Leaf, StateId, TriggerId};

#[derive(Clone, Debug, PartialEq, Eq, Hash)]
pub struct FreeData {
    pub retry_at: u64,
    pub created_on_activation: bool,
}

#[derive(Clone, Debug, Default, PartialEq, Eq, Hash)]
pub struct IdleData {
    pub ticks: u32,
}

#[derive(Clone, Debug, PartialEq, Eq, Hash)]
pub struct Job {
    pub units: u32,
    pub loaded: u32,
}

#[derive(Clone, Debug, PartialEq, Eq, Hash)]
pub struct Transfer {
    pub remaining: u32,
}

#[derive(Default)]
struct Agent {
    chart: hauler::Chart,
    capacity: u32,
    docked: bool,
    log: Vec<String>,
}

impl Delegate for Agent {
    fn chart(&mut self) -> &mut hauler::Chart {
        &mut self.chart
    }

    fn free_data(&mut self, event: Option<&Event>) -> FreeData {
        self.log.push(format!("free_data {:?}", event.map(Event::id)));
        FreeData { retry_at: 0, created_on_activation: event.is_none() && !self.chart.is_active() }
    }

    fn free_capture_retry(&mut self, free: &mut FreeData, at: &u64) {
        free.retry_at = *at;
    }

    fn free_job_fits(&self, _free: &FreeData, job: &Job) -> bool {
        job.units <= self.capacity
    }

    fn free_accept(&mut self, free: &mut FreeData, job: &Job) {
        // Actions run before the source states are exited, so they still see their data.
        self.log.push(format!("accept {} (retry_at {})", job.units, free.retry_at));
    }

    fn idle_capture_tick(&mut self, _free: &mut FreeData, idle: &mut IdleData) {
        idle.ticks += 1;
    }

    fn hauling_data(&mut self, event: Option<&Event>) -> Job {
        match event {
            Some(Event::Reserved { job }) => job.clone(),
            other => panic!("Hauling is only entered by Reserved, got {other:?}"),
        }
    }

    fn hauling_enter(&mut self, hauling: &mut Job) {
        self.log.push(format!("hauling_enter {}", hauling.units));
    }

    fn hauling_exit(&mut self, hauling: &mut Job) {
        self.log.push(format!("hauling_exit {}/{}", hauling.loaded, hauling.units));
    }

    fn to_pickup_docked(&self, _hauling: &Job) -> bool {
        self.docked
    }

    fn loading_data(&mut self, hauling: &mut Job, event: Option<&Event>) -> Transfer {
        // Children get the data of their parents.
        assert_eq!(event.map(Event::id), Some(TriggerId::Tick));
        Transfer { remaining: hauling.units - hauling.loaded }
    }

    fn loading_enter_on_tick(&mut self, _hauling: &mut Job, loading: &mut Transfer) {
        self.log.push(format!("loading_enter_on_tick {}", loading.remaining));
    }

    fn loading_exit(&mut self, hauling: &mut Job, loading: &mut Transfer) {
        self.log.push(format!("loading_exit {} {}", hauling.loaded, loading.remaining));
    }

    fn loading_capture_loaded(&mut self, hauling: &mut Job, loading: &mut Transfer, units: &u32) {
        hauling.loaded += units;
        loading.remaining -= units;
    }

    fn loading_loaded_all(&self, _hauling: &Job, loading: &Transfer) -> bool {
        loading.remaining == 0
    }

    fn loading_empty_job(&self, hauling: &Job, _loading: &Transfer) -> bool {
        hauling.units == 0
    }

    fn on_transition(&mut self, from: Option<StateId>, to: StateId, trigger: Option<TriggerId>) {
        self.log.push(format!("transition {from:?} > {to:?} on {trigger:?}"));
    }
}

fn agent(capacity: u32) -> Agent {
    let mut agent = Agent { capacity, ..Default::default() };
    hauler::activate(&mut agent);
    agent.log.clear();
    agent
}

#[test]
fn activation_creates_the_default_data() {
    let mut agent = Agent::default();
    hauler::activate(&mut agent);

    assert_eq!(agent.log, vec!["free_data None"]);
    assert_eq!(agent.chart.leaf(), Leaf::Idle);
    assert_eq!(agent.chart.free(), Some(&FreeData { retry_at: 0, created_on_activation: true }));
    assert_eq!(agent.chart.idle(), Some(&IdleData { ticks: 0 }));
    assert_eq!(agent.chart.hauling(), None);
    assert_eq!(agent.chart.loading(), None);
}

#[test]
fn callbacks_change_the_data() {
    let mut agent = agent(10);
    hauler::trigger_retry(&mut agent, 42);
    hauler::trigger_tick(&mut agent);
    hauler::trigger_tick(&mut agent);

    assert_eq!(agent.chart.free().map(|f| f.retry_at), Some(42));
    assert_eq!(agent.chart.idle(), Some(&IdleData { ticks: 2 }));

    // The data can also be changed from outside.
    agent.chart.idle_mut().unwrap().ticks = 7;
    assert_eq!(agent.chart.idle(), Some(&IdleData { ticks: 7 }));
}

#[test]
fn a_full_job() {
    let mut agent = agent(10);
    hauler::trigger_retry(&mut agent, 5);

    // Too big: the condition fails and nothing happens.
    hauler::trigger_reserved(&mut agent, Job { units: 20, loaded: 0 });
    assert_eq!(agent.chart.leaf(), Leaf::Idle);
    assert!(agent.log.is_empty());

    hauler::trigger_reserved(&mut agent, Job { units: 3, loaded: 0 });
    assert_eq!(agent.chart.leaf(), Leaf::ToPickup);
    assert_eq!(agent.chart.hauling(), Some(&Job { units: 3, loaded: 0 }));
    // The data of the exited states is gone.
    assert_eq!(agent.chart.free(), None);
    assert_eq!(agent.chart.idle(), None);
    assert_eq!(
        agent.log,
        vec![
            "accept 3 (retry_at 5)",
            "hauling_enter 3",
            "transition Some(Free) > Hauling on Some(Reserved)",
        ]
    );
    agent.log.clear();

    // Not docked yet.
    hauler::trigger_tick(&mut agent);
    assert_eq!(agent.chart.leaf(), Leaf::ToPickup);

    agent.docked = true;
    hauler::trigger_tick(&mut agent);
    assert_eq!(agent.chart.leaf(), Leaf::Loading);
    assert_eq!(agent.chart.loading(), Some(&Transfer { remaining: 3 }));

    hauler::trigger_loaded(&mut agent, 2);
    assert_eq!(agent.chart.hauling(), Some(&Job { units: 3, loaded: 2 }));
    hauler::trigger_tick(&mut agent);
    assert_eq!(agent.chart.leaf(), Leaf::Loading);

    hauler::trigger_loaded(&mut agent, 1);
    hauler::trigger_tick(&mut agent);
    assert_eq!(agent.chart.leaf(), Leaf::Idle);
    assert_eq!(agent.chart.hauling(), None);
    assert_eq!(agent.chart.loading(), None);
    // Free is entered again, with new data.
    assert_eq!(agent.chart.free(), Some(&FreeData { retry_at: 0, created_on_activation: false }));
    assert_eq!(
        agent.log,
        vec![
            "loading_enter_on_tick 3",
            "transition Some(ToPickup) > Loading on Some(Tick)",
            "loading_exit 3 0",
            "hauling_exit 3/3",
            "free_data Some(Tick)",
            "transition Some(Loading) > Idle on Some(Tick)",
        ]
    );
}

#[test]
fn null_transitions_see_the_data() {
    let mut agent = agent(10);
    agent.docked = true;
    hauler::trigger_reserved(&mut agent, Job { units: 0, loaded: 0 });
    agent.log.clear();

    // Entering Loading with an empty job leaves right away.
    hauler::trigger_tick(&mut agent);
    assert_eq!(agent.chart.leaf(), Leaf::Idle);
    assert_eq!(
        agent.log,
        vec![
            "loading_enter_on_tick 0",
            "transition Some(ToPickup) > Loading on Some(Tick)",
            "loading_exit 0 0",
            "hauling_exit 0/0",
            "free_data None",
            "transition Some(Loading) > Idle on None",
        ]
    );
}

#[test]
fn parent_transitions_apply_to_children() {
    let mut agent = agent(10);
    agent.docked = true;
    hauler::trigger_reserved(&mut agent, Job { units: 4, loaded: 0 });
    hauler::trigger_tick(&mut agent);
    assert_eq!(agent.chart.leaf(), Leaf::Loading);

    hauler::trigger_cancel(&mut agent);
    assert_eq!(agent.chart.leaf(), Leaf::Idle);
}

#[test]
fn charts_are_plain_values() {
    let mut agent = agent(10);
    hauler::trigger_reserved(&mut agent, Job { units: 4, loaded: 0 });

    let saved = agent.chart.clone();
    assert_eq!(saved, agent.chart);
    let hash = |chart: &hauler::Chart| {
        use std::hash::{Hash, Hasher};
        let mut hasher = std::collections::hash_map::DefaultHasher::new();
        chart.hash(&mut hasher);
        hasher.finish()
    };
    assert_eq!(hash(&saved), hash(&agent.chart));

    hauler::trigger_cancel(&mut agent);
    assert_ne!(saved, agent.chart);

    // Restoring a saved chart restores the states and their data.
    agent.chart = saved;
    assert_eq!(agent.chart.leaf(), Leaf::ToPickup);
    assert_eq!(agent.chart.hauling(), Some(&Job { units: 4, loaded: 0 }));
    assert!(format!("{:?}", agent.chart).contains("ToPickup"));
}
