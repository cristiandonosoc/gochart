#[path = "default_active.gochart.rs"]
mod door;

use door::{Delegate, Leaf, StateId};

#[derive(Default)]
struct House {
    chart: door::Chart,
    entered: Vec<StateId>,
}

impl Delegate for House {
    fn chart(&mut self) -> &mut door::Chart {
        &mut self.chart
    }

    fn unlatched_enter(&mut self, _shut: &mut u32) {}

    fn ajar_enter(&mut self) {}

    fn on_enter_state(&mut self, state: StateId) {
        self.entered.push(state);
    }
}

#[test]
fn default_charts_are_in_their_default_states() {
    let house = House::default();
    assert!(house.chart.is_active());
    assert_eq!(house.chart.leaf(), Leaf::Latched);
    assert_eq!(house.chart.shut(), Some(&0));
    // No delegate saw the initial states being entered.
    assert!(house.entered.is_empty());
}

#[test]
fn activate_does_nothing_on_an_active_chart() {
    let mut house = House::default();
    *house.chart.shut_mut().unwrap() = 3;
    door::activate(&mut house);
    assert_eq!(house.chart.shut(), Some(&3));
    assert!(house.entered.is_empty());
}

#[test]
fn the_chart_is_inactive_during_callbacks_and_after_deactivate() {
    let mut house = House::default();
    assert!(door::trigger_open(&mut house));
    assert_eq!(house.chart.leaf(), Leaf::Unlatched);
    assert_eq!(house.entered, vec![StateId::Unlatched]);

    door::deactivate(&mut house);
    assert!(!house.chart.is_active());
    assert_eq!(house.chart, door::Chart::INACTIVE);
    door::activate(&mut house);
    assert_eq!(house.chart.leaf(), Leaf::Latched);
}
