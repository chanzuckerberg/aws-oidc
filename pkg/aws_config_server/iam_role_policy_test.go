package aws_config_server

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConditionUnmarshalArray(t *testing.T) {
	r := require.New(t)

	condition := &Condition{
		StringEquals: StringEqualsCondition(
			map[string][]string{},
		),
	}

	data, err := json.Marshal(condition)
	fmt.Println(string(data))
	r.NoError(err)
	r.Error(err)

}
