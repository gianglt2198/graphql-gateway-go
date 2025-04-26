package reflection

import (
	"reflect"

	"github.com/samber/lo"
)

func CallFunctionWithValue[R any](fn any, args ...any) R {
	fnRef := reflect.ValueOf(fn)
	argsVal := lo.Map(args, func(arg any, _ int) reflect.Value {
		return reflect.ValueOf(arg)
	})
	res := fnRef.Call(argsVal)
	if res[0].IsNil() {
		return lo.Empty[R]()
	}
	return res[0].Interface().(R)
}

func CallMethod(methodName string, caller any, args ...any) {
	method := reflect.ValueOf(caller).MethodByName(methodName)
	argVals := lo.Map(args, func(arg any, _ int) reflect.Value {
		return reflect.ValueOf(arg)
	})
	method.Call(argVals)
}

func CallMethodWithValue[R any](methodName string, caller any, args ...any) R {
	method := reflect.ValueOf(caller).MethodByName(methodName)
	argVals := lo.Map(args, func(arg any, _ int) reflect.Value {
		return reflect.ValueOf(arg)
	})
	res := method.Call(argVals)
	return res[0].Interface().(R)
}
