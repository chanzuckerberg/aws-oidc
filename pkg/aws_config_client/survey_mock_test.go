package aws_config_client

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
)

type MockPrompt struct {
	selectResponse []int
	selectIdx      int

	inputResponse []string
	inputIdx      int

	confirmResponse []bool
	confirmIdx      int
}

func (m *MockPrompt) Select(prompt string, options []string, surveyOptions ...survey.AskOpt) (int, error) {
	selectOptions := ""
	for optI, opt := range options {
		optLine := fmt.Sprintf("\n%d:%s", optI, opt)
		selectOptions = selectOptions + optLine
	}
	fmt.Println(selectOptions)
	s := m.selectResponse[m.selectIdx]
	fmt.Println("Selected Option: ", s)
	m.selectIdx++

	return s, nil
}

func (m *MockPrompt) Input(prompt string, defaulted string, surveyOptions ...survey.AskOpt) (string, error) {
	fmt.Println("input prompt: ", prompt)
	i := m.inputResponse[m.inputIdx]
	m.inputIdx++

	// if the input is empty, return the default
	if i == "" {
		fmt.Println("Selected Option: ", defaulted)
		return defaulted, nil
	}
	fmt.Println("Selected Option: ", i)
	return i, nil
}

func (m *MockPrompt) Confirm(prompt string, defaulted bool, surveyOptions ...survey.AskOpt) (bool, error) {
	fmt.Println("confirm prompt: ", prompt)
	c := m.confirmResponse[m.confirmIdx]
	m.confirmIdx++
	fmt.Println("Selected Option: ", c)
	return c, nil
}
