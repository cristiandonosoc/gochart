#[path = "delegate_defaults.gochart.rs"]
mod worker;

use worker::{Event, Leaf};

/// Sends Reserved only: it implements the data of Busy and nothing else.
struct Apply<'a>(&'a mut worker::Chart);

impl worker::Delegate for Apply<'_> {
    fn chart(&mut self) -> &mut worker::Chart {
        self.0
    }

    fn busy_data(&mut self, event: Option<&Event>) -> u32 {
        let Some(Event::Reserved { job }) = event else { unreachable!("Busy is entered on Reserved") };
        *job
    }
}

/// Sends Tick: it implements the conditions and actions and leaves the data out.
#[derive(Default)]
struct Think {
    chart: worker::Chart,
    bored: bool,
    wandered: u32,
}

impl worker::Delegate for Think {
    fn chart(&mut self) -> &mut worker::Chart {
        &mut self.chart
    }

    fn idle_bored(&self) -> bool {
        self.bored
    }

    fn idle_wander(&mut self) {
        self.wandered += 1;
    }

    fn busy_done(&self, busy: &u32) -> bool {
        *busy == 0
    }
}

#[test]
fn delegates_implement_only_what_they_reach() {
    let mut think = Think::default();
    worker::activate(&mut think);
    assert!(!worker::trigger_tick(&mut think));

    assert!(worker::trigger_reserved(&mut Apply(&mut think.chart), 7));
    assert_eq!(think.chart.leaf(), Leaf::Busy);
    assert_eq!(think.chart.busy(), Some(&7));

    *think.chart.busy_mut().unwrap() = 0;
    assert!(worker::trigger_tick(&mut think));
    think.bored = true;
    assert!(worker::trigger_tick(&mut think));
    assert_eq!(think.chart.leaf(), Leaf::Wandering);
    assert_eq!(think.wandered, 1);
}

#[test]
fn a_missing_callback_panics_with_its_name() {
    let mut chart = worker::Chart::default();
    worker::activate(&mut Apply(&mut chart));
    let result = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        worker::trigger_tick(&mut Apply(&mut chart));
    }));
    let panic = result.expect_err("Apply doesn't implement idle_bored");
    let message = match panic.downcast_ref::<String>() {
        Some(message) => message.clone(),
        None => panic.downcast_ref::<&str>().map(|s| s.to_string()).unwrap_or_default(),
    };
    assert!(message.contains("gochart Worker: idle_bored is not implemented by this delegate"), "{message}");
}
