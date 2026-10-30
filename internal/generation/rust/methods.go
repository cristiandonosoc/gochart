package rust

import (
	"fmt"
	"strings"
)

// delegateMethod is a method of the generated Delegate trait that the user implements.
type delegateMethod struct {
	Name string
	Doc  string
	// Mutable is whether the method takes |&mut self| (actions) or |&self| (conditions).
	Mutable bool
	// Params are the parameters after self, already formatted ("job: &HaulJob").
	Params []string
	Return string

	// signature identifies the method, so that the same callback can be named in several places
	// (eg. the same condition for two transitions of a state on the same trigger).
	signature string
}

func (dm *delegateMethod) Declaration() string {
	return dm.Signature() + ";"
}

// Signature is the declaration without the final ";", to give the method a body.
func (dm *delegateMethod) Signature() string {
	self := "&self"
	if dm.Mutable {
		self = "&mut self"
	}
	params := append([]string{self}, dm.Params...)
	ret := ""
	if dm.Return != "" {
		ret = " -> " + dm.Return
	}
	return fmt.Sprintf("fn %s(%s)%s", dm.Name, strings.Join(params, ", "), ret)
}

// kReservedMethods are the methods the trait always has.
var kReservedMethods = map[string]struct{}{
	"chart":          {},
	"on_enter_state": {},
	"on_exit_state":  {},
	"on_transition":  {},
}

// pathParams returns the parameters that receive the data of |states|. |taken| are names already
// used by other parameters of the same method.
func pathParams(states []*rustState, mutable bool, taken []string) []string {
	ref := "&"
	if mutable {
		ref = "&mut "
	}

	used := map[string]struct{}{}
	for _, name := range taken {
		used[name] = struct{}{}
	}

	params := make([]string, 0, len(states))
	for _, state := range states {
		name := state.Snake
		if _, ok := used[name]; ok {
			name += "_data"
		}
		used[name] = struct{}{}
		params = append(params, fmt.Sprintf("%s: %s%s", name, ref, state.DataType))
	}
	return params
}

func argParams(trigger *rustTrigger) []string {
	if trigger == nil {
		return nil
	}
	params := make([]string, 0, len(trigger.Args))
	for _, arg := range trigger.Args {
		params = append(params, fmt.Sprintf("%s: &%s", arg.Name, arg.Type))
	}
	return params
}

func argNames(trigger *rustTrigger) []string {
	if trigger == nil {
		return nil
	}
	names := make([]string, 0, len(trigger.Args))
	for _, arg := range trigger.Args {
		names = append(names, arg.Name)
	}
	return names
}

func (m *model) addMethod(dm *delegateMethod) error {
	if _, ok := kReservedMethods[dm.Name]; ok {
		return fmt.Errorf("callback %q (%s) collides with a method that every delegate has", dm.Name, dm.Doc)
	}

	dm.signature = dm.Declaration()
	for _, other := range m.Methods {
		if other.Name != dm.Name {
			continue
		}
		if other.signature == dm.signature {
			// Same callback used in several places.
			return nil
		}
		return fmt.Errorf("callback %q is needed with two different signatures (%s; %s), rename one of them",
			dm.Name, other.Doc, dm.Doc)
	}

	m.Methods = append(m.Methods, dm)
	return nil
}

// Callback names. These are used both for the trait declaration and the calls.

func dataMethodName(s *rustState) string { return s.Snake + "_data" }

func reactionMethodName(s *rustState, kind string, trigger *rustTrigger) string {
	if trigger == nil {
		return fmt.Sprintf("%s_%s", s.Snake, kind)
	}
	return fmt.Sprintf("%s_%s_on_%s", s.Snake, kind, trigger.Snake)
}

func captureMethodName(s *rustState, trigger *rustTrigger) string {
	return fmt.Sprintf("%s_capture_%s", s.Snake, trigger.Snake)
}

func callbackMethodName(s *rustState, callback string) string {
	return fmt.Sprintf("%s_%s", s.Snake, callback)
}

func (m *model) stateLabel(s *rustState) string {
	if s.isRoot {
		return fmt.Sprintf("gochart %s", m.gc.Name)
	}
	return fmt.Sprintf("state %s", s.Name)
}

func (m *model) collectMethods() error {
	for _, s := range m.AllStates() {
		label := m.stateLabel(s)

		// Data constructor.
		if s.HasData() && !s.DataDefault {
			params := pathParams(s.AncestorDataPath(), true, []string{"event"})
			params = append(params, "event: Option<&Event>")
			if err := m.addMethod(&delegateMethod{
				Name:    dataMethodName(s),
				Doc:     fmt.Sprintf("Creates the data of %s when it is entered. |event| is the trigger that caused the entry, if any.", label),
				Mutable: true,
				Params:  params,
				Return:  s.DataType,
			}); err != nil {
				return err
			}
		}

		// Reactions.
		for _, kind := range []string{"enter", "exit"} {
			reactions := s.ir.EnterReactions
			if kind == "exit" {
				reactions = s.ir.ExitReactions
			}
			for _, reaction := range reactions {
				trigger := m.triggerMap[reaction.Trigger]
				doc := fmt.Sprintf("`%s` reaction of %s.", kind, label)
				if trigger != nil {
					doc = fmt.Sprintf("`%s %s` reaction of %s.", kind, trigger.Name, label)
				}
				params := pathParams(s.DataPath(), true, argNames(trigger))
				params = append(params, argParams(trigger)...)
				if err := m.addMethod(&delegateMethod{
					Name:    reactionMethodName(s, kind, trigger),
					Doc:     doc,
					Mutable: true,
					Params:  params,
				}); err != nil {
					return err
				}
			}
		}

		// Captures.
		for _, capture := range s.ir.Captures {
			trigger := m.triggerMap[capture.Trigger]
			params := pathParams(s.DataPath(), true, argNames(trigger))
			params = append(params, argParams(trigger)...)
			if err := m.addMethod(&delegateMethod{
				Name:    captureMethodName(s, trigger),
				Doc:     fmt.Sprintf("`capture %s` of %s.", trigger.Name, label),
				Mutable: true,
				Params:  params,
			}); err != nil {
				return err
			}
			if capture.ConditionCallback != "" {
				if err := m.addConditionMethod(s, trigger, capture.ConditionCallback,
					fmt.Sprintf("Condition of `capture %s` of %s.", trigger.Name, label)); err != nil {
					return err
				}
			}
		}

		// Transitions.
		for _, transition := range s.ir.Transitions {
			trigger := m.triggerMap[transition.Trigger]
			triggerName := "null"
			if trigger != nil {
				triggerName = trigger.Name
			}
			what := fmt.Sprintf("`transition %s > %s` of %s", triggerName, transition.To.Name, label)

			if transition.ConditionCallback != "" {
				if err := m.addConditionMethod(s, trigger, transition.ConditionCallback,
					fmt.Sprintf("Condition of %s.", what)); err != nil {
					return err
				}
			}
			if transition.ActionCallback != "" {
				params := pathParams(s.DataPath(), true, argNames(trigger))
				params = append(params, argParams(trigger)...)
				if err := m.addMethod(&delegateMethod{
					Name:    callbackMethodName(s, transition.ActionCallback),
					Doc:     fmt.Sprintf("Action of %s. Runs before the states are exited.", what),
					Mutable: true,
					Params:  params,
				}); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (m *model) addConditionMethod(s *rustState, trigger *rustTrigger, callback, doc string) error {
	params := pathParams(s.DataPath(), false, argNames(trigger))
	params = append(params, argParams(trigger)...)
	return m.addMethod(&delegateMethod{
		Name:    callbackMethodName(s, callback),
		Doc:     doc,
		Mutable: false,
		Params:  params,
		Return:  "bool",
	})
}
