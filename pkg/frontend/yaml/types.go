package yaml

type YamlStatechart struct {
	Name        string            `yaml:"name"`
	Triggers    []*YamlTrigger    `yaml:"triggers"`
	States      []*YamlState      `yaml:"states"`
	Transitions []*YamlTransition `yaml:"transitions"`
}

type YamlTrigger struct {
	Name       string `yaml:"name"`
	ArgsString string `yaml:"arguments_string"`
}

type YamlState struct {
	Name   string `yaml:"name"`
	Parent string `yaml:"parent"`

	Options []*YamlStateOption `yaml:"options"`

	EnterReactions []string `yaml:"enter_reactions"`
	ExitReactions  []string `yaml:"exit_reactions"`
}

type YamlStateOption struct {
	Name string `yaml:"name"`

	IntValue    string `yaml:"int_value"`
	StringValue string `yaml:"string_value"`
	BoolValue   string `yaml:"bool_value"`
}

type YamlTransition struct {
	From    string `yaml:"from"`
	To      string `yaml:"to"`
	Trigger string `yaml:"trigger"`
}
