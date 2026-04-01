package quiz

// yamlQuestion is the raw YAML struct for a single question.
// Release 3+ fields (answers, choices, media) are parsed but not validated.
type yamlQuestion struct {
	Text    string        `yaml:"text"`
	Answer  string        `yaml:"answer"`
	Answers []string      `yaml:"answers"`
	Choices []interface{} `yaml:"choices"`
	Media   interface{}   `yaml:"media"`
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
