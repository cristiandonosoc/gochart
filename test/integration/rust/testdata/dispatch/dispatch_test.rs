#[path = "dispatch.gochart.rs"]
mod dispatch;

use dispatch::{Delegate, Leaf, StateId, TriggerId};

#[derive(Default)]
struct Pinger {
    chart: dispatch::Chart,
    choice: u32,
    log: Vec<String>,
}

impl Pinger {
    fn take_log(&mut self) -> Vec<String> {
        std::mem::take(&mut self.log)
    }
}

impl Delegate for Pinger {
    fn chart(&mut self) -> &mut dispatch::Chart {
        &mut self.chart
    }

    fn gochart_capture_ping(&mut self, value: &u32) {
        self.log.push(format!("gochart {value}"));
    }

    fn outer_capture_ping(&mut self, value: &u32) {
        self.log.push(format!("outer {value}"));
    }

    fn outer_big(&self, value: &u32) -> bool {
        *value >= 100
    }

    fn inner_capture_ping(&mut self, value: &u32) {
        self.log.push(format!("inner {value}"));
    }

    fn inner_small(&self, value: &u32) -> bool {
        *value < 10
    }

    fn inner_first_choice(&self) -> bool {
        self.choice == 1
    }

    fn inner_second_choice(&self) -> bool {
        self.choice == 2
    }

    fn on_enter_state(&mut self, state: StateId) {
        self.log.push(format!("+{}", state.name()));
    }

    fn on_exit_state(&mut self, state: StateId) {
        self.log.push(format!("-{}", state.name()));
    }

    fn on_transition(&mut self, from: Option<StateId>, to: StateId, trigger: Option<TriggerId>) {
        let from = from.map_or("gochart", StateId::name);
        let trigger = trigger.map_or("null", TriggerId::name);
        self.log.push(format!("{from} > {} on {trigger}", to.name()));
    }
}

#[test]
fn triggers_bubble_up_until_handled() {
    let mut p = Pinger::default();
    dispatch::activate(&mut p);
    assert_eq!(p.take_log(), vec!["+Outer", "+Inner"]);

    dispatch::trigger_ping(&mut p, 1);
    dispatch::trigger_ping(&mut p, 50);
    dispatch::trigger_ping(&mut p, 200);
    assert_eq!(p.take_log(), vec!["inner 1", "gochart 50", "outer 200"]);
}

#[test]
fn entries_are_evaluated_in_order() {
    for (choice, leaf) in [(1, Leaf::Second), (2, Leaf::Third), (3, Leaf::Fallback)] {
        let mut p = Pinger { choice, ..Default::default() };
        dispatch::activate(&mut p);
        p.take_log();

        dispatch::trigger_go(&mut p);
        assert_eq!(p.chart.leaf(), leaf);
        assert_eq!(
            p.take_log(),
            vec![
                "-Inner".to_string(),
                "-Outer".to_string(),
                format!("+{}", leaf.name()),
                format!("Inner > {} on Go", leaf.name()),
            ]
        );

        dispatch::trigger_back(&mut p);
        assert_eq!(p.chart.leaf(), Leaf::Inner);
    }
}

#[test]
fn unhandled_triggers_do_nothing() {
    let mut p = Pinger::default();
    dispatch::activate(&mut p);
    p.take_log();

    dispatch::trigger_back(&mut p);
    assert_eq!(p.chart.leaf(), Leaf::Inner);
    assert!(p.log.is_empty());
}

#[test]
fn the_chart_is_taken_out_during_callbacks() {
    struct Peeker {
        chart: dispatch::Chart,
        saw_active: Option<bool>,
    }

    impl Delegate for Peeker {
        fn chart(&mut self) -> &mut dispatch::Chart {
            &mut self.chart
        }
        fn gochart_capture_ping(&mut self, _value: &u32) {}
        fn outer_capture_ping(&mut self, _value: &u32) {}
        fn outer_big(&self, _value: &u32) -> bool {
            false
        }
        fn inner_capture_ping(&mut self, _value: &u32) {
            self.saw_active = Some(self.chart.is_active());
        }
        fn inner_small(&self, _value: &u32) -> bool {
            true
        }
        fn inner_first_choice(&self) -> bool {
            false
        }
        fn inner_second_choice(&self) -> bool {
            false
        }
    }

    let mut p = Peeker { chart: dispatch::Chart::default(), saw_active: None };
    dispatch::activate(&mut p);
    dispatch::trigger_ping(&mut p, 1);
    assert_eq!(p.saw_active, Some(false));
    assert!(p.chart.is_active());
}
