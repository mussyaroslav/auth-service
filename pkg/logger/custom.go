package logger

import (
	"log/slog"
	"reflect"
)

func Err(err error) slog.Attr {
	if err == nil {
		return slog.Attr{
			Key:   "no errors",
			Value: slog.StringValue(""),
		}
	}
	return slog.Attr{
		Key:   "ERROR",
		Value: slog.StringValue(err.Error()),
	}
}

func Obj(obj any) slog.Attr {
	return slog.Attr{
		Key:   reflect.TypeOf(obj).String(),
		Value: slog.AnyValue(obj),
	}
}
