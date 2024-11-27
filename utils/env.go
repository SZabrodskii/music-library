package utils

import (
	"os"
	"reflect"
	"strconv"
	"strings"
)

func GetEnv[T comparable](key string, defaultValue T) T {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	defaultValueType := reflect.TypeOf(defaultValue)

	switch defaultValueType.Kind() {
	case reflect.String:
		return interface{}(value).(T)
	case reflect.Slice:
		valueSlice := strings.Split(value, ",")
		sliceType := defaultValueType.Elem()
		slice := reflect.MakeSlice(reflect.SliceOf(sliceType), len(valueSlice), len(valueSlice))
		for i, v := range valueSlice {
			sliceValue := convert(v, sliceType)
			slice.Index(i).Set(reflect.ValueOf(sliceValue))
		}
		return slice.Interface().(T)
	default:
		return convert(value, defaultValueType).Interface().(T)
	}
}

func convert(value string, targetType reflect.Type) reflect.Value {
	switch targetType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if parsedValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return reflect.ValueOf(parsedValue).Convert(targetType)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if parsedValue, err := strconv.ParseUint(value, 10, 64); err == nil {
			return reflect.ValueOf(parsedValue).Convert(targetType)
		}

	case reflect.Bool:
		if parsedValue, err := strconv.ParseBool(value); err == nil {
			return reflect.ValueOf(parsedValue).Convert(targetType)
		}

	case reflect.String:
		return reflect.ValueOf(value).Convert(targetType)

	default:
		return reflect.Zero(targetType)
	}

	return reflect.Zero(targetType)
}
