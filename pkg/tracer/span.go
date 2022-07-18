package tracer

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

var translatedSpan = map[string]string{}

const (
	folderSep = "/"
	funcSep   = "."
	newSep    = ":"
)

func StartSpan(ctx context.Context, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if traceTracer == nil {
		return ctx, nil
	}

	spanName, err := getFuncName(2)
	if err != nil {
		handleErr(err, "err get func name")
	}

	newName, ok := translatedSpan[spanName]
	if ok {
		spanName = newName
	} else if strings.Contains(spanName, folderSep) {
		runeFunc := strings.Split(spanName, folderSep)
		if len(runeFunc) > 1 {
			runeFunc = runeFunc[1:]
		}

		lastRunePoint := len(runeFunc) - 1
		lastRune := runeFunc[lastRunePoint]
		if strings.Contains(lastRune, funcSep) {
			runeFunc = append(runeFunc[:lastRunePoint], strings.Split(lastRune, funcSep)...)
		}

		newSpanName := strings.Join(runeFunc, newSep)
		translatedSpan[spanName] = newSpanName
		spanName = newSpanName
	}

	fmt.Println(spanName)

	return traceTracer.Start(ctx, spanName, opts...)
}

func getFuncName(level int) (string, error) {
	pc, _, _, ok := runtime.Caller(level)
	if !ok {
		return "", errors.New("err from runtime.caller")
	}

	details := runtime.FuncForPC(pc)
	if details == nil {
		return "", errors.New("err from runtime.FuncForPC")
	}

	return details.Name(), nil
}
