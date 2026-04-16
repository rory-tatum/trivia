package quiz

// yamlQuestion is the raw YAML struct for a single question.
type yamlQuestion struct {
	Text    string     `yaml:"text"`
	Answer  string     `yaml:"answer"`
	Answers []string   `yaml:"answers"`
	Choices []string   `yaml:"choices"`
	Media   *yamlMedia `yaml:"media"`
}

// yamlMedia is the raw YAML struct for a media attachment.
type yamlMedia struct {
	Type string `yaml:"type"`
	URL  string `yaml:"url"`
}

// yamlRound is the raw YAML struct for a single round.
type yamlRound struct {
	Name      string         `yaml:"name"`
	Questions []yamlQuestion `yaml:"questions"`
}

// yamlQuiz is the raw YAML struct for the top-level quiz file.
type yamlQuiz struct {
	Title  string      `yaml:"title"`
	Rounds []yamlRound `yaml:"rounds"`
}
