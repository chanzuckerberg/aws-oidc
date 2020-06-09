package aws_config_client

import "github.com/AlecAivazis/survey/v2"

type MockPrompt struct {
	selectResponse []int
	selectIdx      int

	inputResponse []string
	inputIdx      int

	confirmResponse []bool
	confirmIdx      int
}

func (m *MockPrompt) Select(prompt string, options []string, surveyOptions ...survey.AskOpt) (int, error) {
	s := m.selectResponse[m.selectIdx]
	m.selectIdx++

	return s, nil
}

func (m *MockPrompt) Input(prompt string, defaulted string, surveyOptions ...survey.AskOpt) (string, error) {
	i := m.inputResponse[m.inputIdx]
	m.inputIdx++

	// if the input is empty, return the default
	if i == "" {
		return defaulted, nil
	}

	return i, nil
}

func (m *MockPrompt) Confirm(prompt string, defaulted bool, surveyOptions ...survey.AskOpt) (bool, error) {
	c := m.confirmResponse[m.confirmIdx]
	m.confirmIdx++

	return c, nil
}
