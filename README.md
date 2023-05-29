# go-json-validator

**Go-json-validator is now available as part of [go-ejson](https://github.com/galdor/go-ejson).**

## Introduction
The go-json-validator library contains tooling to validate JSON data. While
the standard Go parser will only check basic types, this package lets
developers write code to validate any property.

## Usage
Refer to the [Go package
documentation](https://pkg.go.dev/github.com/galdor/go-json-validator) for
information about the API.

### Example
```go
type SignupData struct {
	EmailAddress         string `json:"emailAddress"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password"`
	AcceptTOS            bool   `json:"acceptTOS"`
}

func (d *SignupData) ValidateJSON(v *jsonvalidator.Validator) {
	v.CheckStringMatch2("emailAddress", d.EmailAddress, EmailAddressRE,
		"invalidEmailAddress", "invalid email address")

	v.CheckStringLengthMin("password", d.Password, 8)

	v.Check("passwordConfirmation", d.PasswordConfirmation == d.Password,
		"passwordMismatch", "password confirmation and password do not match")

	v.Check("acceptTOS", d.AcceptTOS, "tosNotAccepted",
		"you must accept terms of services to sign up")
}
```

Simply implement `ValidateJSON` for structures that need extra validation, and
use `jsonvalidation.Unmarshal` to both decode and validate data.

Errors are reported using type `ValidationErrors` which can contain one or
more validation errors. Each validation error contains a JSON pointer to the
location of the error and an error code to facilitate error handling.

# Licensing
Go-json-validator is open source software distributed under the
[ISC](https://opensource.org/licenses/ISC) license.
