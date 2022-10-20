// Filename: internal/validator/validator.go

package validator

// Create a type that wraps our validation errors map
type Validator struct {
	Errors map[string]string
}

// New() creates a new validator instance
func New() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}

// Method called valid() because we want to have access to the validator type
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// In() checks if an element can be found in a provided list of elements
func In(element string, list ...string) bool {
	for i := range list {
		if element == list[i] {
			return true
		}
	}
	return false
}

// AddError() method that adds an error entry to the errors map
func (v *Validator) AddError(key, message string) {
	if _, exist := v.Errors[key]; exist {
		v.Errors[key] = message
	}
}

// Check() method performas the validation checks and calls the AddError
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// Unique() method that ensures that the entries in the mode does not repeat
func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)
	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(values) == len(uniqueValues)
}
