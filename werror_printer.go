package werror

import (
	"bytes"
	"fmt"
	"sort"
)

func GenerateErrorString(err error, outputEveryCallingStack bool) string {
	if zerror, ok := err.(*werror); ok {
		return generateZerrorString(zerror, outputEveryCallingStack)
	}
	if fancy, ok := err.(fmt.Formatter); ok {
		// This is a rich error type, like those produced by github.com/pkg/errors.
		return fmt.Sprintf("%+v", fancy)
	}
	return err.Error()
}

func generateZerrorString(err *werror, outputEveryCallingStack bool) string {
	var buffer bytes.Buffer
	writeMessage(err, &buffer)
	writeParams(err, &buffer)
	writeCause(err, &buffer, outputEveryCallingStack)
	writeStack(err, &buffer, outputEveryCallingStack)
	return buffer.String()
}

func writeMessage(err *werror, buffer *bytes.Buffer) {
	if err.message == "" {
		return
	}
	buffer.WriteString(err.message)
}

func writeParams(err *werror, buffer *bytes.Buffer) {
	safeParams := err.SafeParams()
	var safeKeys []string
	for k := range safeParams {
		safeKeys = append(safeKeys, k)
	}
	sort.Strings(safeKeys)
	messageAndParams := err.message != "" && len(safeParams) != 0
	messageOrParams := err.message != "" || len(safeParams) != 0
	if messageAndParams {
		buffer.WriteString(" ")
	}
	for _, safeKey := range safeKeys {
		buffer.WriteString(fmt.Sprintf("%+v:%+v", safeKey, safeParams[safeKey]))
		// If it is not the last param, add a separator
		if !(safeKeys[len(safeKeys)-1] == safeKey) {
			buffer.WriteString(", ")
		}
	}
	if messageOrParams {
		buffer.WriteString("\n")
	}
}

func writeCause(err *werror, buffer *bytes.Buffer, outputEveryCallingStack bool) {
	if err.cause != nil {
		buffer.WriteString(GenerateErrorString(err.cause, outputEveryCallingStack))
	}
}

func writeStack(err *werror, buffer *bytes.Buffer, outputEveryCallingStack bool) {
	if _, ok := err.cause.(*werror); ok {
		if !outputEveryCallingStack {
			return
		}
	}
	buffer.WriteString(fmt.Sprintf("%+v", err.stack))
}
