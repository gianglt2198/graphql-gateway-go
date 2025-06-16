package reflection

import (
	"errors"
	"fmt"
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

func CallMethodWithError[R any](methodName string, caller any, args ...any) (R, error) {
	method := reflect.ValueOf(caller).MethodByName(methodName)
	argsVal := lo.Map(args, func(arg any, _ int) reflect.Value {
		return reflect.ValueOf(arg)
	})
	res := method.Call(argsVal)
	if res[1].IsNil() {
		return res[0].Interface().(R), nil
	}
	return lo.Empty[R](), res[1].Interface().(error)
}

func SetField(obj any, name string, value any) error {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		return fmt.Errorf("no such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if !val.Type().AssignableTo(structFieldType) {
		return errors.New("provided value type not assignable to obj field type")
	}

	structFieldValue.Set(val)
	return nil
}
