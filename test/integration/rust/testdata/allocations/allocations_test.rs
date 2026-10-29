#[path = "allocations.gochart.rs"]
mod allocations;

use std::alloc::{GlobalAlloc, Layout, System};
use std::cell::Cell;

use allocations::{Delegate, Event, Leaf};

// Counts the allocations of the current thread, so other test threads don't interfere.
struct CountingAllocator;

thread_local! {
    static ALLOCATIONS: Cell<usize> = const { Cell::new(0) };
}

unsafe impl GlobalAlloc for CountingAllocator {
    unsafe fn alloc(&self, layout: Layout) -> *mut u8 {
        let _ = ALLOCATIONS.try_with(|count| count.set(count.get() + 1));
        System.alloc(layout)
    }

    unsafe fn dealloc(&self, ptr: *mut u8, layout: Layout) {
        System.dealloc(ptr, layout)
    }
}

#[global_allocator]
static ALLOCATOR: CountingAllocator = CountingAllocator;

fn allocations() -> usize {
    ALLOCATIONS.with(Cell::get)
}

#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct Counter {
    pub ticks: u32,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Job {
    pub remaining: u32,
}

#[derive(Default)]
struct Worker {
    chart: allocations::Chart,
}

impl Delegate for Worker {
    fn chart(&mut self) -> &mut allocations::Chart {
        &mut self.chart
    }

    fn idle_capture_tick(&mut self, idle: &mut Counter, _dt: &u32) {
        idle.ticks += 1;
    }

    fn working_data(&mut self, event: Option<&Event>) -> Job {
        match event {
            Some(Event::Start { units }) => Job { remaining: *units },
            _ => Job { remaining: 0 },
        }
    }

    fn busy_finished(&self, working: &Job, dt: &u32) -> bool {
        working.remaining <= *dt
    }

    fn clock_capture_tick(&mut self, clock: &mut Counter, _dt: &u32) {
        clock.ticks += 1;
    }
}

#[test]
fn the_allocator_counts() {
    let before = allocations();
    let boxed = std::hint::black_box(Box::new(1u64));
    assert_eq!(allocations() - before, 1);
    drop(boxed);
}

#[test]
fn running_a_chart_does_not_allocate() {
    let mut worker = Worker::default();

    let before = allocations();
    allocations::activate(&mut worker);
    for _ in 0..100 {
        allocations::trigger_tick(&mut worker, 1);
        allocations::trigger_start(&mut worker, 3);
        allocations::trigger_tick(&mut worker, 1);
        allocations::trigger_tick(&mut worker, 5);
        allocations::trigger_start(&mut worker, 3);
        allocations::trigger_stop(&mut worker);
    }
    let leaves = worker.chart.leaves();
    let clock = worker.chart.clock().cloned();
    allocations::deactivate(&mut worker);
    let after = allocations();

    assert_eq!(after - before, 0);
    assert_eq!(leaves[1], Some(Leaf::Idle));
    assert_eq!(clock, Some(Counter { ticks: 300 }));
}

#[test]
fn chart_size() {
    // The chart is a plain value: its size is the size of the data of the states that can be
    // active at the same time, plus discriminants.
    let size = std::mem::size_of::<allocations::Chart>();
    assert!(size <= 32, "chart is {size} bytes");
}
