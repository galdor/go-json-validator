package jsonvalidator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestFoo struct {
	String string
	Bar    *TestBar
	Bars   []*TestBar
}

type TestBar struct {
	Integers []int
}

func (foo *TestFoo) ValidateJSON(v *Validator) {
	v.CheckStringLengthMin("String", foo.String, 3)

	v.CheckOptionalObject("Bar", foo.Bar)

	v.WithChild("Bars", func() {
		for i, bar := range foo.Bars {
			v.CheckObject(i, bar)
		}
	})
}

func (bar *TestBar) ValidateJSON(v *Validator) {
	v.WithChild("Integers", func() {
		for i, integer := range bar.Integers {
			v.CheckIntMax(i, integer, 10)
		}
	})
}

func TestValidate(t *testing.T) {
	assert := assert.New(t)

	var data TestFoo
	var err error
	var validationErrs ValidationErrors
	var validationErr *ValidationError

	// Valid data
	data = TestFoo{
		String: "abcdef",
		Bar:    &TestBar{Integers: []int{1, 2, 3}},
		Bars: []*TestBar{
			{Integers: []int{4}},
			{Integers: []int{5, 6}},
		},
	}

	assert.NoError(Validate(&data))

	// Valid data with null optional object
	data = TestFoo{
		String: "abcdef",
		Bar:    nil,
	}

	assert.NoError(Validate(&data))

	// Simple top-level violation
	data = TestFoo{
		String: "ab",
		Bar:    nil,
	}

	err = Validate(&data)

	if assert.ErrorAs(err, &validationErrs) {
		if assert.Equal(1, len(validationErrs)) {
			validationErr = validationErrs[0]
			assert.Equal("/String", validationErr.Pointer.String())
			assert.Equal("stringTooShort", validationErr.Code)
		}
	}

	// Null objects in an object array
	data = TestFoo{
		String: "abcdef",
		Bars: []*TestBar{
			{Integers: []int{4}},
			nil,
			{Integers: []int{5, 6}},
			nil,
		},
	}

	err = Validate(&data)

	if assert.ErrorAs(err, &validationErrs) {
		if assert.Equal(2, len(validationErrs)) {
			validationErr = validationErrs[0]
			assert.Equal("/Bars/1", validationErr.Pointer.String())
			assert.Equal("missingValue", validationErr.Code)

			validationErr = validationErrs[1]
			assert.Equal("/Bars/3", validationErr.Pointer.String())
			assert.Equal("missingValue", validationErr.Code)
		}
	}

	// Nested violations
	data = TestFoo{
		String: "abcdef",
		Bars: []*TestBar{
			nil,
			{Integers: []int{15}},
			{Integers: []int{5, 20}},
		},
	}

	err = Validate(&data)

	if assert.ErrorAs(err, &validationErrs) {
		if assert.Equal(3, len(validationErrs)) {
			validationErr = validationErrs[0]
			assert.Equal("/Bars/0", validationErr.Pointer.String())
			assert.Equal("missingValue", validationErr.Code)

			validationErr = validationErrs[1]
			assert.Equal("/Bars/1/Integers/0", validationErr.Pointer.String())
			assert.Equal("integerTooLarge", validationErr.Code)

			validationErr = validationErrs[2]
			assert.Equal("/Bars/2/Integers/1", validationErr.Pointer.String())
			assert.Equal("integerTooLarge", validationErr.Code)
		}
	}

	// Invalid top-level type
	err = Unmarshal([]byte(`42`), &data)

	if assert.ErrorAs(err, &validationErrs) {
		if assert.Equal(1, len(validationErrs)) {
			validationErr = validationErrs[0]
			assert.Equal("", validationErr.Pointer.String())
			assert.Equal("invalidValueType", validationErr.Code)
		}
	}

	// Invalid member type
	err = Unmarshal([]byte(`{"String": 42}`), &data)

	if assert.ErrorAs(err, &validationErrs) {
		if assert.Equal(1, len(validationErrs)) {
			validationErr = validationErrs[0]
			assert.Equal("/String", validationErr.Pointer.String())
			assert.Equal("invalidValueType", validationErr.Code)
		}
	}

	// Invalid nested member type
	//
	// The standard JSON parser returns the error on the array, nothing we can
	// do about it.
	err = Unmarshal([]byte(`{"String": "abcd", "Bars": [{"Integers": true}]}`),
		&data)

	if assert.ErrorAs(err, &validationErrs) {
		if assert.Equal(1, len(validationErrs)) {
			validationErr = validationErrs[0]
			assert.Equal("/Bars/Integers", validationErr.Pointer.String())
			assert.Equal("invalidValueType", validationErr.Code)
		}
	}
}
