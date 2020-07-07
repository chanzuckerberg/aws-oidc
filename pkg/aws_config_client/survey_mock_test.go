package aws_config_client

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/go-errors/errors"
)

type MockPrompt struct {
	inputs   []interface{}
	inputIdx int
}

func (m *MockPrompt) Select(prompt string, options []string, surveyOptions ...survey.AskOpt) (int, error) {
	fmt.Println("\ninputIdx: ", m.inputIdx)
	selectOptions := ""
	for optI, opt := range options {
		optLine := fmt.Sprintf("\n%d:%s", optI, opt)
		selectOptions = selectOptions + optLine
	}
	fmt.Println(selectOptions)

	selection, ok := m.inputs[m.inputIdx].(int)
	if !ok {
		return -1, errors.Errorf("Expected int input, got %T type for %v value", m.inputs[m.inputIdx], m.inputs[m.inputIdx])
	}

	fmt.Printf("Selection Index: %d, value: %s", selection, options[selection])
	m.inputIdx++
	return selection, nil
}

func (m *MockPrompt) Input(prompt string, defaulted string, surveyOptions ...survey.AskOpt) (string, error) {
	fmt.Println("\ninputIdx: ", m.inputIdx)
	textInput, ok := m.inputs[m.inputIdx].(string)
	if !ok {
		return "", errors.Errorf("Expected string input, got %T type for %v value", m.inputs[m.inputIdx], m.inputs[m.inputIdx])
	}
	// if the input is empty, return the default
	if textInput == "" {
		textInput = defaulted
	}

	fmt.Printf("Input value: %s\n", textInput)
	m.inputIdx++
	return textInput, nil
}

func (m *MockPrompt) Confirm(prompt string, defaulted bool, surveyOptions ...survey.AskOpt) (bool, error) {
	fmt.Println("\ninputIdx: ", m.inputIdx)
	confirmation, ok := m.inputs[m.inputIdx].(bool)
	if !ok {
		return defaulted, errors.Errorf("Expected bool input, got %T type for %v value", m.inputs[m.inputIdx], m.inputs[m.inputIdx])
	}

	fmt.Printf("Confirm value: %v\n", confirmation)
	m.inputIdx++
	return confirmation, nil
}
