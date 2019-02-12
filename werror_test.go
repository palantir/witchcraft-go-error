package werror_test

import (
	"fmt"
	"testing"

	"github.com/palantir/witchcraft-go-error"
	"github.com/palantir/witchcraft-go-error/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type paramObject struct {
	A string
	B int
	c string
}

type testParameterStorerObject struct {
	safeParams   map[string]interface{}
	unsafeParams map[string]interface{}
}

func (t testParameterStorerObject) SafeParams() map[string]interface{} {
	return t.safeParams
}

func (t testParameterStorerObject) UnsafeParams() map[string]interface{} {
	return t.unsafeParams
}

const pkgPath = "github.com/palantir/witchcraft-go-error"

func TestError_Format(t *testing.T) {
	for _, currCase := range []struct {
		name               string
		err                error
		stringified        string
		verbose            string
		extraVerboseRegexp string
	}{
		{
			name: "new error without safe params",
			err: werror.Error("message",
				werror.UnsafeParam("unsafeKey", "unsafeKey"),
			),
			stringified: "message",
			verbose:     `message`,
			extraVerboseRegexp: `^message
` + pkgPath + `_test.TestError_Format
	.+
testing.tRunner
	.+
runtime.goexit
	.+$`,
		},
		{
			name: "new error with params",
			err: werror.Error("message",
				werror.UnsafeParam("unsafeKey", "unsafeKey"),
				werror.SafeParam("safeKey", paramObject{
					A: "public value A",
					B: 1,
					c: "private value C",
				}),
			),
			stringified: "message",
			verbose:     `message map[safeKey:{A:public value A B:1 c:private value C}]`,
			extraVerboseRegexp: `^message map\[safeKey:{A:public value A B:1 c:private value C}\]
` + pkgPath + `_test.TestError_Format
	.+
testing.tRunner
	.+
runtime.goexit
	.+$`,
		},
		{
			name: "error wrapped with errors and werror",
			err: werror.Wrap(
				errors.Wrap(
					werror.Error(
						"root cause",
						werror.UnsafeParam("unsafeRootKey", "unsafeRootValue"),
						werror.SafeParam("safeRootKey", "safeRootValue"),
					),
					"first wrapper",
				),
				"second wrapper",
				werror.UnsafeParam("unsafeWrapperKey", "unsafeWrapperValue"),
				werror.SafeParam("safeWrapperKey", "safeWrapperValue"),
			),
			stringified: "second wrapper: first wrapper: root cause",
			// Unfortunately losing params from errors wrapped by "errors" package,
			// because %v gets turned into %s in the "errors" formatting code.
			verbose: `second wrapper map[safeWrapperKey:safeWrapperValue]: first wrapper: root cause`,
			extraVerboseRegexp: `^root cause map\[safeRootKey:safeRootValue\]
` + pkgPath + `_test.TestError_Format
	.+
testing.tRunner
	.+
runtime.goexit
	.+
first wrapper
` + pkgPath + `_test.TestError_Format
	.+
testing.tRunner
	.+
runtime.goexit
	.+
second wrapper map\[safeWrapperKey:safeWrapperValue\]
` + pkgPath + `_test.TestError_Format
	.+
testing.tRunner
	.+
runtime.goexit
	.+$`,
		},
		{
			name: "wrapped with empty string",
			err: werror.Wrap(
				werror.Error("rootcause"),
				"",
			),
			stringified: "rootcause",
			verbose:     `rootcause`,
			extraVerboseRegexp: `^rootcause
` + pkgPath + `_test.TestError_Format
	.+
testing.tRunner
	.+
runtime.goexit
	.+

` + pkgPath + `_test.TestError_Format
	.+
testing.tRunner
	.+
runtime.goexit
	.+$`,
		},
		{
			name: "wrapped with empty string and params",
			err: werror.Wrap(
				werror.Error("rootcause"),
				"",
				werror.SafeParam("safeEmptyWrapperKey", "safeEmptyWrapperValue"),
				werror.UnsafeParam("unsafeWrapperKey", "unsafeWrapperValue"),
			),
			stringified: "rootcause",
			verbose:     `map[safeEmptyWrapperKey:safeEmptyWrapperValue]: rootcause`,
			extraVerboseRegexp: `^rootcause
` + pkgPath + `_test.TestError_Format
	.+
testing.tRunner
	.+
runtime.goexit
	.+
map\[safeEmptyWrapperKey:safeEmptyWrapperValue\]
` + pkgPath + `_test.TestError_Format
	.+
testing.tRunner
	.+
runtime.goexit
	.+$`,
		},
		{
			name: "wrapped custom error with params",
			err: werror.Wrap(
				fmt.Errorf("customErr"),
				"wrapper",
				werror.SafeParam("safeWrapperKey", "safeWrapperValue"),
				werror.UnsafeParam("unsafeWrapperKey", "unsafeWrapperValue"),
			),
			stringified: "wrapper: customErr",
			verbose:     `wrapper map[safeWrapperKey:safeWrapperValue]: customErr`,
			extraVerboseRegexp: `^customErr
wrapper map\[safeWrapperKey:safeWrapperValue\]
` + pkgPath + `_test.TestError_Format
	.+
testing.tRunner
	.+
runtime.goexit
	.+$`,
		},
		{
			name: "converted custom error",
			err: werror.Convert(
				fmt.Errorf("customErr"),
			),
			stringified: "customErr",
			verbose:     `customErr`,
			extraVerboseRegexp: `^customErr

` + pkgPath + `_test.TestError_Format
	.+
testing.tRunner
	.+
runtime.goexit
	.+$`,
		},
	} {
		t.Run(currCase.name, func(t *testing.T) {
			require.Error(t, currCase.err)
			assert.Equal(t, currCase.stringified, fmt.Sprintf("%s", currCase.err))
			assert.Equal(t, currCase.verbose, fmt.Sprintf("%v", currCase.err))
			assert.Regexp(t, currCase.extraVerboseRegexp, fmt.Sprintf("%+v", currCase.err))
		})
	}
}

func TestParamsFromError(t *testing.T) {
	for _, currCase := range []struct {
		name             string
		err              error
		wantSafeParams   map[string]interface{}
		wantUnsafeParams map[string]interface{}
	}{
		{
			name:             "without params",
			err:              fmt.Errorf("regular error"),
			wantSafeParams:   map[string]interface{}{},
			wantUnsafeParams: map[string]interface{}{},
		},
		{
			name:             "nil error",
			err:              nil,
			wantSafeParams:   map[string]interface{}{},
			wantUnsafeParams: map[string]interface{}{},
		},
		{
			name: "with the same safe and unsafe param key",
			err: werror.Error("err",
				werror.SafeParam("key", "safeValue"),
				werror.UnsafeParam("key", "unsafeValue")),
			wantSafeParams: map[string]interface{}{},
			wantUnsafeParams: map[string]interface{}{
				"key": "unsafeValue",
			},
		},
		{
			name: "with the same safe and unsafe param key (reverse order)",
			err: werror.Error("err",
				werror.UnsafeParam("key", "unsafeValue"),
				werror.SafeParam("key", "safeValue")),
			wantSafeParams: map[string]interface{}{
				"key": "safeValue",
			},
			wantUnsafeParams: map[string]interface{}{},
		},
		{
			name: "with nested params",
			err: werror.Wrap(
				errors.Wrap(
					werror.Error(
						"root cause",
						werror.UnsafeParam("unsafeRootKey", "unsafeRootValue"),
						werror.SafeParam("safeRootKey", "safeRootValue"),
					),
					"first wrapper",
				),
				"second wrapper",
				werror.UnsafeParam("unsafeWrapperKey", "unsafeWrapperValue"),
				werror.SafeParam("safeWrapperKey", "safeWrapperValue"),
			),
			wantSafeParams: map[string]interface{}{
				"safeRootKey":    "safeRootValue",
				"safeWrapperKey": "safeWrapperValue",
			},
			wantUnsafeParams: map[string]interface{}{
				"unsafeRootKey":    "unsafeRootValue",
				"unsafeWrapperKey": "unsafeWrapperValue",
			},
		},
		{
			name: "with empty safe and unsafe params param",
			err: werror.Error("error",
				werror.SafeAndUnsafeParams(
					map[string]interface{}{},
					map[string]interface{}{},
				),
			),
			wantSafeParams:   map[string]interface{}{},
			wantUnsafeParams: map[string]interface{}{},
		},
		{
			name: "with safe and unsafe params param",
			err: werror.Error("error",
				werror.SafeAndUnsafeParams(
					map[string]interface{}{
						"safeKey": "safeVal",
						"config":  "logging",
					},
					map[string]interface{}{
						"unsafeKey": "unsafeVal",
						"commonKey": "level4",
						"fileName":  "logger.txt",
					},
				),
			),
			wantSafeParams: map[string]interface{}{
				"safeKey": "safeVal",
				"config":  "logging",
			},
			wantUnsafeParams: map[string]interface{}{
				"unsafeKey": "unsafeVal",
				"commonKey": "level4",
				"fileName":  "logger.txt",
			},
		},
	} {
		t.Run(currCase.name, func(t *testing.T) {
			gotSafeParams, gotUnsafeParams := werror.ParamsFromError(currCase.err)
			assert.Equal(t, currCase.wantSafeParams, gotSafeParams)
			assert.Equal(t, currCase.wantUnsafeParams, gotUnsafeParams)
		})
	}
}

func TestParamFromError(t *testing.T) {
	for _, currCase := range []struct {
		name          string
		err           error
		expectedValue interface{}
		expectedSafe  bool
	}{{
		name: "nil error",
		err:  nil,
	}, {
		name: "error without param",
		err:  werror.Error("err"),
	}, {
		name: "error with safe param",
		err: werror.Error("err",
			werror.SafeParam("key", "value")),
		expectedValue: "value",
		expectedSafe:  true,
	}, {
		name: "error with unsafe param",
		err: werror.Error("err",
			werror.UnsafeParam("key", "value")),
		expectedValue: "value",
		expectedSafe:  false,
	}, {
		name: "error with duplicated param",
		err: werror.Error("err",
			werror.UnsafeParam("key", "value1"),
			werror.SafeParam("key", "value2"),
			werror.SafeParam("key", "value3"),
		),
		expectedValue: "value3",
		expectedSafe:  true,
	}} {
		t.Run(currCase.name, func(t *testing.T) {
			gotValue, gotSafe := werror.ParamFromError(currCase.err, "key")
			assert.Equal(t, currCase.expectedValue, gotValue)
			assert.Equal(t, currCase.expectedSafe, gotSafe)
		})
	}
}

func TestParamsFromError_FromParameterStorerObject(t *testing.T) {
	for _, currCase := range []struct {
		name string
		//parameterStorerObject werror.ParamStorer
		inErr            error
		wantSafeParams   map[string]interface{}
		wantUnsafeParams map[string]interface{}
	}{
		{
			name: "empty parameterStorer",
			inErr: werror.Error(
				"error",
				werror.Params(testParameterStorerObject{}),
			),
			wantSafeParams:   map[string]interface{}{},
			wantUnsafeParams: map[string]interface{}{},
		},
		{
			name: "parameterStorer with safe and unsafe params",
			inErr: werror.Error(
				"error",
				werror.Params(testParameterStorerObject{
					safeParams: map[string]interface{}{
						"safeObjectParamKey": "safeObjectParamValue",
					},
					unsafeParams: map[string]interface{}{
						"unsafeObjectParamKey": "unsafeObjectParamValue",
					},
				}),
			),
			wantSafeParams: map[string]interface{}{
				"safeObjectParamKey": "safeObjectParamValue",
			},
			wantUnsafeParams: map[string]interface{}{
				"unsafeObjectParamKey": "unsafeObjectParamValue",
			},
		},
		{
			name: "non-werror ParamStorer error",
			inErr: &customParamStorerError{
				msg: "error",
				safeParams: map[string]interface{}{
					"safeObjectParamKey": "safeObjectParamValue",
				},
				unsafeParams: map[string]interface{}{
					"unsafeObjectParamKey": "unsafeObjectParamValue",
				},
			},
			wantSafeParams: map[string]interface{}{
				"safeObjectParamKey": "safeObjectParamValue",
			},
			wantUnsafeParams: map[string]interface{}{
				"unsafeObjectParamKey": "unsafeObjectParamValue",
			},
		},
		{
			name: "werror with non-werror ParamStorer error cause",
			inErr: werror.Wrap(
				&customParamStorerError{
					msg: "error",
					safeParams: map[string]interface{}{
						"safeObjectParamKey": "safeObjectParamValue",
					},
					unsafeParams: map[string]interface{}{
						"unsafeObjectParamKey": "unsafeObjectParamValue",
					},
				},
				"error",
			),
			wantSafeParams: map[string]interface{}{
				"safeObjectParamKey": "safeObjectParamValue",
			},
			wantUnsafeParams: map[string]interface{}{
				"unsafeObjectParamKey": "unsafeObjectParamValue",
			},
		},
	} {
		t.Run(currCase.name, func(t *testing.T) {
			gotSafeParams, gotUnsafeParams := werror.ParamsFromError(currCase.inErr)
			assert.Equal(t, currCase.wantSafeParams, gotSafeParams)
			assert.Equal(t, currCase.wantUnsafeParams, gotUnsafeParams)
		})
	}
}

type customParamStorerError struct {
	msg                      string
	safeParams, unsafeParams map[string]interface{}
}

func (e *customParamStorerError) SafeParams() map[string]interface{} {
	return e.safeParams
}

func (e *customParamStorerError) UnsafeParams() map[string]interface{} {
	return e.unsafeParams
}

func (e *customParamStorerError) Error() string {
	return fmt.Sprintf("customError: %s, safeParams: %v, unsafeParams: %v", e.msg, e.safeParams, e.unsafeParams)
}

func TestConvert(t *testing.T) {
	for _, currCase := range []struct {
		name string
		err  error
	}{{
		name: "nil error",
		err:  nil,
	}, {
		name: "custom error",
		err:  fmt.Errorf("custom error"),
	}, {
		name: "werror error",
		err:  werror.Error("werror error"),
	}, {
		name: "wrapped custom error",
		err:  werror.Wrap(fmt.Errorf("custom error"), "wrapped error"),
	}} {
		t.Run(currCase.name, func(t *testing.T) {
			if currCase.err == nil {
				assert.Nil(t, werror.Convert(currCase.err))
			} else {
				converted := werror.Convert(currCase.err)
				assert.EqualError(t, converted, currCase.err.Error(), "should have the same message")
				assert.Contains(t, fmt.Sprintf("%+v", converted), "TestConvert", "contains stacktrace")
				assert.Equal(t, converted, werror.Convert(converted), "should be idempotent")
			}
		})
	}
}

func TestWrap_NilErrorIsNil(t *testing.T) {
	require.Nil(t, werror.Wrap(nil, "<-- nil!"), "werror.Wrap(nil) was not nil")
}

func TestRootCause(t *testing.T) {
	werrorErr := werror.Error("werror err")
	customErr := fmt.Errorf("custom err")
	for _, currCase := range []struct {
		name      string
		rootCause error
		err       error
	}{{
		name:      "nil error",
		rootCause: nil,
		err:       nil,
	}, {
		name:      "new werror error",
		rootCause: werrorErr,
		err:       werrorErr,
	}, {
		name:      "wrapped werror error",
		rootCause: werrorErr,
		err:       werror.Wrap(werrorErr, "wrap", werror.SafeParam("safeKey", "safeVal")),
	}, {
		name:      "converted werror error",
		rootCause: werrorErr,
		err:       werror.Convert(werrorErr),
	}, {
		name:      "custom error",
		rootCause: customErr,
		err:       customErr,
	}, {
		name:      "wrapped custom error",
		rootCause: customErr,
		err:       werror.Wrap(customErr, "wrap", werror.SafeParam("safeKey", "safeVal")),
	}, {
		name:      "converted custom error",
		rootCause: customErr,
		err:       werror.Convert(customErr),
	}} {
		t.Run(currCase.name, func(t *testing.T) {
			assert.Equal(t, currCase.rootCause, werror.RootCause(currCase.err))
		})
	}
}
