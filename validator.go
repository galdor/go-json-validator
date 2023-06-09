package jsonvalidator

import (
	"bytes"
	"fmt"
	"net/url"
	"reflect"
	"regexp"

	jsonpointer "github.com/galdor/go-json-pointer"
)

type ValidationError struct {
	Pointer jsonpointer.Pointer `json:"pointer"`
	Code    string              `json:"code"`
	Message string              `json:"message"`
}

type ValidationErrors []*ValidationError

type Validator struct {
	Pointer jsonpointer.Pointer
	Errors  ValidationErrors
}

type Validatable interface {
	ValidateJSON(v *Validator)
}

func (err ValidationError) Error() string {
	if len(err.Pointer) == 0 {
		return err.Message
	} else {
		return fmt.Sprintf("%v: %s", err.Pointer, err.Message)
	}
}

func (errs ValidationErrors) Error() string {
	var buf bytes.Buffer

	buf.WriteString("invalid data:")

	for _, err := range errs {
		buf.WriteString("\n  ")
		buf.WriteString(err.Error())
	}

	return buf.String()
}

func Validate(value interface{}) error {
	v := NewValidator()

	if validatableValue, ok := value.(Validatable); ok {
		validatableValue.ValidateJSON(v)
	}

	if len(v.Errors) > 0 {
		return v.Error()
	}

	return nil
}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Error() ValidationErrors {
	if len(v.Errors) == 0 {
		return nil
	}

	return v.Errors
}

func (v *Validator) Push(token interface{}) {
	v.Pointer = v.Pointer.Child(token)
}

func (v *Validator) Pop() {
	v.Pointer = v.Pointer.Parent()
}

func (v *Validator) WithChild(token interface{}, fn func()) {
	v.Push(token)
	defer v.Pop()

	fn()
}

func (v *Validator) AddError(token interface{}, code, format string, args ...interface{}) {
	pointer := v.Pointer.Child(token)

	err := ValidationError{
		Pointer: pointer,
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}

	v.Errors = append(v.Errors, &err)
}

func (v *Validator) Check(token interface{}, value bool, code, format string, args ...interface{}) bool {
	if !value {
		v.AddError(token, code, format, args...)
	}

	return value
}

func (v *Validator) CheckIntMin(token interface{}, i int, min int) bool {
	return v.Check(token, i >= min, "integerTooSmall",
		"integer must be greater or equal to %d", min)
}

func (v *Validator) CheckIntMax(token interface{}, i int, max int) bool {
	return v.Check(token, i <= max, "integerTooLarge",
		"integer must be lower or equal to %d", max)
}

func (v *Validator) CheckIntMinMax(token interface{}, i int, min, max int) bool {
	if !v.CheckIntMin(token, i, min) {
		return false
	}

	return v.CheckIntMax(token, i, max)
}

func (v *Validator) CheckStringLengthMin(token interface{}, s string, min int) bool {
	return v.Check(token, len(s) >= min, "stringTooShort",
		"string length must be greater or equal to %d", min)
}

func (v *Validator) CheckStringLengthMax(token interface{}, s string, max int) bool {
	return v.Check(token, len(s) <= max, "stringTooLong",
		"string length must be lower or equal to %d", max)
}

func (v *Validator) CheckStringLengthMinMax(token interface{}, s string, min, max int) bool {
	if !v.CheckStringLengthMin(token, s, min) {
		return false
	}

	return v.CheckStringLengthMax(token, s, max)
}

func (v *Validator) CheckStringNotEmpty(token interface{}, s string) bool {
	return v.Check(token, s != "", "missingOrEmptyString",
		"missing or empty string")
}

func (v *Validator) CheckStringMatch(token interface{}, s string, re *regexp.Regexp) bool {
	return v.CheckStringMatch2(token, s, re, "invalidStringFormat",
		"string must match the following regular expression: %s",
		re.String())
}

func (v *Validator) CheckStringMatch2(token interface{}, s string, re *regexp.Regexp, code, format string, args ...interface{}) bool {
	if !re.MatchString(s) {
		v.AddError(token, code, format, args...)
		return false
	}

	return true
}

func (v *Validator) CheckStringURI(token interface{}, s string) bool {
	// The url.Parse function parses URI references. Most of the time we are
	// interested in URIs, so we check that there is a schema.

	uri, err := url.Parse(s)
	if err != nil {
		v.AddError(token, "invalidURIFormat", "string must be a valid uri")
		return false
	}

	if uri.Scheme == "" {
		v.AddError(token, "missingURIScheme", "uri must have a scheme")
		return false
	}

	return true
}

func (v *Validator) CheckOptionalObject(token interface{}, value interface{}) bool {
	if !checkObject(value) {
		return true
	}

	return v.doCheckObject(token, value)
}

func (v *Validator) CheckObject(token interface{}, value interface{}) bool {
	if !checkObject(value) {
		v.AddError(token, "missingValue", "missing value")
		return false
	}

	return v.doCheckObject(token, value)
}

func (v *Validator) doCheckObject(token interface{}, value interface{}) bool {
	nbErrors := len(v.Errors)

	value2, ok := value.(Validatable)
	if !ok {
		return true
	}

	v.Push(token)
	value2.ValidateJSON(v)
	v.Pop()

	return len(v.Errors) == nbErrors
}

func checkObject(value interface{}) bool {
	valueType := reflect.TypeOf(value)
	if valueType == nil {
		return false
	}

	if valueType.Kind() != reflect.Pointer {
		panic(fmt.Sprintf("value %#v (%T) is not a pointer", value, value))
	}

	pointedValueType := valueType.Elem()
	if pointedValueType.Kind() != reflect.Struct {
		panic(fmt.Sprintf("value %#v (%T) is not a pointer to a structure",
			value, value))
	}

	return !reflect.ValueOf(value).IsZero()
}
