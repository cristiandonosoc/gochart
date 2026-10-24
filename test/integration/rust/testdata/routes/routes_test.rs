#[path = "routes.gochart.rs"]
mod routes;

use routes::{Delegate, Leaf, StateId};

#[derive(Default)]
struct Recorder {
    chart: routes::Chart,
    calls: Vec<&'static str>,
}

macro_rules! record {
    ($($name:ident),* $(,)?) => {
        $(fn $name(&mut self) { self.calls.push(stringify!($name)); })*
    };
}

impl Delegate for Recorder {
    fn chart(&mut self) -> &mut routes::Chart {
        &mut self.chart
    }

    record!(
        gochart_enter, gochart_exit,
        first_enter, first_exit, first_son1_enter, first_son1_exit, first_son2_enter, first_son2_exit,
        second_enter, second_exit, second_son1_enter, second_son1_exit,
        third_enter, third_exit,
        deep_enter, deep_exit, deep_a_enter, deep_a_exit, deep_a1_enter, deep_a1_exit, deep_b_enter, deep_b_exit,
    );
}

impl Recorder {
    fn take_calls(&mut self) -> Vec<&'static str> {
        std::mem::take(&mut self.calls)
    }
}

#[test]
fn same_routes_as_cpp() {
    let mut d = Recorder::default();
    assert!(!d.chart.is_active());

    routes::activate(&mut d);
    assert_eq!(d.chart.leaf(), Leaf::FirstSon2);

    routes::trigger_go_to_second(&mut d);
    assert_eq!(d.chart.leaf(), Leaf::SecondSon1);

    routes::trigger_go_to_first(&mut d);
    assert_eq!(d.chart.leaf(), Leaf::FirstSon1);

    routes::trigger_go_to_third(&mut d);
    assert_eq!(d.chart.leaf(), Leaf::Third);

    routes::trigger_go_to_first(&mut d);
    assert_eq!(d.chart.leaf(), Leaf::FirstSon2);

    routes::deactivate(&mut d);
    assert!(!d.chart.is_active());

    assert_eq!(
        d.take_calls(),
        vec![
            "gochart_enter",
            "first_enter",
            "first_son2_enter",
            //
            "first_son2_exit",
            "first_exit",
            "second_enter",
            "second_son1_enter",
            //
            "second_son1_exit",
            "second_exit",
            "first_enter",
            "first_son1_enter",
            //
            "first_son1_exit",
            "first_exit",
            "third_enter",
            //
            "third_exit",
            "first_enter",
            "first_son2_enter",
            //
            "first_son2_exit",
            "first_exit",
            "gochart_exit",
        ]
    );
}

#[test]
fn routes_between_different_depths_stay_in_the_common_parent() {
    let mut d = Recorder::default();
    routes::activate(&mut d);
    routes::trigger_go_to_deep(&mut d);
    assert_eq!(d.chart.leaf(), Leaf::DeepA1);
    assert!(d.chart.is_in(StateId::Deep));
    assert!(d.chart.is_in(StateId::DeepA));
    assert!(!d.chart.is_in(StateId::DeepB));
    d.take_calls();

    routes::trigger_hop(&mut d);
    assert_eq!(d.chart.leaf(), Leaf::DeepB);
    assert_eq!(d.take_calls(), vec!["deep_a1_exit", "deep_a_exit", "deep_b_enter"]);
}

#[test]
fn transitions_into_a_child_keep_the_source_active() {
    let mut d = Recorder::default();
    routes::activate(&mut d);
    routes::trigger_go_to_deep(&mut d);
    d.take_calls();

    routes::trigger_sideways(&mut d);
    assert_eq!(d.chart.leaf(), Leaf::DeepB);
    assert_eq!(d.take_calls(), vec!["deep_a1_exit", "deep_a_exit", "deep_b_enter"]);
}

#[test]
fn self_transition_exits_and_enters() {
    let mut d = Recorder::default();
    routes::activate(&mut d);
    routes::trigger_go_to_deep(&mut d);
    d.take_calls();

    routes::trigger_again(&mut d);
    assert_eq!(d.chart.leaf(), Leaf::DeepA1);
    assert_eq!(d.take_calls(), vec!["deep_a1_exit", "deep_a1_enter"]);
}

#[test]
fn transition_to_an_ancestor_reenters_it() {
    let mut d = Recorder::default();
    routes::activate(&mut d);
    routes::trigger_go_to_deep(&mut d);
    d.take_calls();

    routes::trigger_up(&mut d);
    assert_eq!(d.chart.leaf(), Leaf::DeepA1);
    assert_eq!(d.take_calls(), vec!["deep_a1_exit", "deep_a_exit", "deep_a_enter", "deep_a1_enter"]);
}

#[test]
fn gochart_level_transition_to_the_current_state_reenters_it() {
    let mut d = Recorder::default();
    routes::activate(&mut d);
    d.take_calls();

    // Root transitions exit the whole active branch, like the C++ backend.
    routes::trigger_go_to_first(&mut d);
    assert_eq!(d.chart.leaf(), Leaf::FirstSon2);
    assert_eq!(d.take_calls(), vec!["first_son2_exit", "first_exit", "first_enter", "first_son2_enter"]);
}

#[test]
fn triggers_are_ignored_while_inactive() {
    let mut d = Recorder::default();
    routes::trigger_go_to_second(&mut d);
    assert!(!d.chart.is_active());
    assert!(d.calls.is_empty());
}

#[test]
fn state_metadata() {
    assert_eq!(StateId::DeepA1.parent(), Some(StateId::DeepA));
    assert_eq!(StateId::Deep.parent(), None);
    assert_eq!(StateId::DeepA1.name(), "DeepA1");
    assert_eq!(StateId::DeepA.as_leaf(), None);
    assert_eq!(StateId::DeepA1.as_leaf(), Some(Leaf::DeepA1));
    assert_eq!(StateId::from(Leaf::Third), StateId::Third);
    assert_eq!(routes::TriggerId::Hop.name(), "Hop");
    assert_eq!(routes::Event::Hop.id(), routes::TriggerId::Hop);
    assert_eq!(StateId::ALL.len(), 10);
    assert_eq!(Leaf::ALL.len(), 6);
    assert_eq!(routes::REGION_COUNT, 1);
}
