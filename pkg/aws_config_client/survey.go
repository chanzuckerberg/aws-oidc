package aws_config_client

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/pkg/errors"
)

type Prompt interface {
	// Select surveys the user for an answer
	// returns the int corresponding to the slice's selected element
	Select(prompt string, options []string, surveyOptions ...survey.AskOpt) (int, error)
	// Input will ask the user for a free-form input
	// with a default option
	Input(prompt string, defaulted string, surveyOptions ...survey.AskOpt) (string, error)
	// Confirm will confirm with the user with a defaulted option
	Confirm(prompt string, defaulted bool, surveryOptions ...survey.AskOpt) (bool, error)
}

type Survey struct{}

func (s *Survey) Select(prompt string, options []string, surveyOptions ...survey.AskOpt) (int, error) {
	var chosen string
	err := survey.AskOne(
		&survey.Select{
			Message: prompt,
			Options: options,
		},
		&chosen,
		surveyOptions...,
	)
	if err != nil {
		return 0, errors.Wrap(err, "error asking user to select")
	}

	for idx, option := range options {
		if chosen == option {
			return idx, nil
		}
	}
	return 0, errors.Errorf("selected option (%s) not expected", chosen)
}

func (s *Survey) Input(prompt string, defaulted string, surveyOptions ...survey.AskOpt) (string, error) {
	var input string

	err := survey.AskOne(
		&survey.Input{
			Message: prompt,
			Default: defaulted,
		},
		&input,
		surveyOptions...,
	)
	if err != nil {
		return "", errors.Wrap(err, "error asking user for input")
	}

	return input, nil
}

func (s *Survey) Confirm(prompt string, defaulted bool, surveryOptions ...survey.AskOpt) (bool, error) {
	var answer bool

	err := survey.AskOne(
		&survey.Confirm{
			Message: prompt,
			Default: defaulted,
		},
		&answer,
		surveryOptions...,
	)
	if err != nil {
		return false, errors.Wrap(err, "error asking for confirmation")
	}
	return answer, nil
}
