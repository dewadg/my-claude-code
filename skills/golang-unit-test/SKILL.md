---
name: golang-unit-test
description: Write Golang unit test functions in test table format, with capability to set before/after hook. ALWAYS use when the user asks to write unit test.
effort: medium
---

Write Golang unit test functions in test table format, with capability to set before/after hook.

## Input
Function/method/its signature to write the unit tests for AND the list of scenario that are going to be covered.

## Output
Unit test functions in table-driven format.

### Output during implementation
```
## Writing unit test: <function/method name>

Working on task 3/7: <scenario>
[...implementation happening...]
✓ Scenario complete

Working on task 4/7: <scenario>
[...implementation happening...]
✓ Scenario complete
```

### Output when completed
```
## Unit tests completed

**Source:** <file-name>: <function/method name>
**Test file:** <test-file-name>: <test function name>
**Progress:** 7/7 scenario covered ✓

### Completed This Session
- [x] Scenario 1
- [x] Scenario 2
...
```

## Steps

1. **Evaluate**
    - Check for existing unit test functions, if it's not in table driven then rewrite
    - Check for existing mocking libraries

2. **Go through each scenario**
    - Each scenario represent a table row in the table-driven tests
    - Setup the mock accordingly
    - Write the tests in format as defined below

3. **Validate**
    - Ensure the tests can be run
    
## Unit Test Function Format

### Struct method
```go
func Test{STRUCT_NAME}_{METHOD_NAME}(t *testing.T) {
    type fields struct {
        // Struct fields here.

        // IMPORTANT: For fields with interface types, replace the type with mock types if exist
    }
	
    type args struct {
		// Function arguments
    }
	
    tests := []struct {
        name    string
        fields  fields
        args    args
        want    {METHOD RETURN TYPE}
        wantErr assert.ErrorAssertionFunc // Omit if the method doesn't throw error
		before  func(fields, args) // Setup the mocks here
		after   func(fields, args) // Cleanup the mocks here
    }{
    // TODO: Add test cases.
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            u := &{STRUCT_NAME}{
               // Pass tt.fields to the struct here 
            }
            got, err := u.{METHOD_NAME}(tt.args.ctx)
            if !tt.wantErr(t, err, fmt.Sprintf("{METHOD_NAME}(%v)", tt.args.ctx)) {
                return
            }
            assert.Equalf(t, tt.want, got, "{METHOD_NAME}(%v)", tt.args.ctx)
		})
	}
}
```

### Function
```go
func Test{FUNCTION_NAME}(t *testing.T) {
	type args struct {
        // Function arguments here
	}
	tests := []struct {
		name string
		args args
		want {FUNCTION RETURN TYPE}
        wantErr assert.ErrorAssertionFunc // Omit if the function doesn't throw error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := {FUNCTION_NAME}(tt.args.amount); got != tt.want {
				t.Errorf("{FUNCTION_NAME}() = %v, want %v", got, tt.want)
			}
		})
	}
}
```
