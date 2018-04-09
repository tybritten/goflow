package functions

import (
	"github.com/nyaruka/goflow/excellent/types"
	"github.com/nyaruka/goflow/utils"
	"strings"
)

// ErrorPassThru wraps an XFunction and returns immediately if any arg is an error
func ErrorPassThru(f XFunction) XFunction {
	return func(env utils.Environment, args ...types.XValue) types.XValue {
		for _, arg := range args {
			if types.IsError(arg) {
				return arg
			}
		}
		return f(env, args...)
	}
}

// ArgCountCheck wraps an XFunction and checks the number of args
func ArgCountCheck(name string, count int, f XFunction) XFunction {
	return func(env utils.Environment, args ...types.XValue) types.XValue {
		if len(args) != count {
			return types.NewXErrorf("%s takes %d argument(s), got %d", strings.ToUpper(name), count, len(args))
		}
		return f(env, args...)
	}
}

// OneStringFunction creates an XFunction from a single string function
func OneStringFunction(name string, f func(utils.Environment, types.XString) types.XValue) XFunction {
	return ErrorPassThru(ArgCountCheck(name, 1, func(env utils.Environment, args ...types.XValue) types.XValue {
		return f(env, types.ToXString(args[0]))
	}))
}

// TwoStringFunction creates an XFunction from a function that takes two strings
func TwoStringFunction(name string, f func(utils.Environment, types.XString, types.XString) types.XValue) XFunction {
	return ErrorPassThru(ArgCountCheck(name, 2, func(env utils.Environment, args ...types.XValue) types.XValue {
		return f(env, types.ToXString(args[0]), types.ToXString(args[1]))
	}))
}

// StringAndIntegerFunction creates an XFunction from a function that takes a string and an integer
func StringAndIntegerFunction(name string, f func(utils.Environment, types.XString, int) types.XValue) XFunction {
	return ErrorPassThru(ArgCountCheck(name, 2, func(env utils.Environment, args ...types.XValue) types.XValue {
		num, err := types.ToInteger(args[1])
		if err != nil {
			return types.NewXError(err)
		}

		return f(env, types.ToXString(args[0]), num)
	}))
}

// OneNumberFunction creates an XFunction from a single number function
func OneNumberFunction(name string, f func(utils.Environment, types.XNumber) types.XValue) XFunction {
	return ErrorPassThru(ArgCountCheck(name, 1, func(env utils.Environment, args ...types.XValue) types.XValue {
		num, err := types.ToXNumber(args[0])
		if err != nil {
			return types.NewXError(err)
		}

		return f(env, num)
	}))
}

// TwoNumberFunction creates an XFunction from a function that takes two numbers
func TwoNumberFunction(name string, f func(utils.Environment, types.XNumber, types.XNumber) types.XValue) XFunction {
	return ErrorPassThru(ArgCountCheck(name, 2, func(env utils.Environment, args ...types.XValue) types.XValue {
		num1, err := types.ToXNumber(args[0])
		if err != nil {
			return types.NewXError(err)
		}
		num2, err := types.ToXNumber(args[1])
		if err != nil {
			return types.NewXError(err)
		}

		return f(env, num1, num2)
	}))
}

// OneDateFunction creates an XFunction from a single number function
func OneDateFunction(name string, f func(utils.Environment, types.XTime) types.XValue) XFunction {
	return ErrorPassThru(ArgCountCheck(name, 1, func(env utils.Environment, args ...types.XValue) types.XValue {
		date, err := types.ToXTime(env, args[0])
		if err != nil {
			return types.NewXError(err)
		}

		return f(env, date)
	}))
}
