package types_test

import (
	"fmt"
	"testing"

	"github.com/nyaruka/goflow/excellent/types"

	"github.com/stretchr/testify/assert"
)

func TestCompare(t *testing.T) {
	var tests = []struct {
		x1       types.XValue
		x2       types.XValue
		result   int
		hasError bool
	}{
		{nil, nil, 0, false},
		{nil, types.NewXText(""), 0, true},
		{types.NewXError(fmt.Errorf("Error")), types.NewXError(fmt.Errorf("Error")), 0, false},
		{types.NewXError(fmt.Errorf("Error")), types.XDateTimeZero, 0, true}, // type mismatch
		{types.NewXText("bob"), types.NewXText("bob"), 0, false},
		{types.NewXText("bob"), types.NewXText("cat"), -1, false},
		{types.NewXText("bob"), types.NewXText("ann"), 1, false},
		{types.NewXNumberFromInt(123), types.NewXNumberFromInt(123), 0, false},
		{types.NewXNumberFromInt(123), types.NewXNumberFromInt(124), -1, false},
		{types.NewXNumberFromInt(123), types.NewXNumberFromInt(122), 1, false},
	}

	for _, test := range tests {
		result, err := types.Compare(test.x1, test.x2)

		if test.hasError {
			assert.Error(t, err, "expected error for inputs '%s' and '%s'", test.x1, test.x2)
		} else {
			assert.NoError(t, err, "unexpected error for inputs '%s' and '%s'", test.x1, test.x2)
			assert.Equal(t, test.result, result, "result mismatch for inputs '%s' and '%s'", test.x1, test.x2)
		}
	}
}
