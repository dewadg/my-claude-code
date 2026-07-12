---
name: golang-unit-test
description: >
  Write Go unit tests in one canonical table-driven format, with optional before/after hooks per
  case. ALWAYS use when the user asks to write, add, or rewrite a Go unit test — e.g. "write unit
  tests for this handler", "cover this method", "add a test case for the error path". Also use when
  an existing Go test is not table-driven and needs to be brought into the standard shape. Not for
  integration or HTTP endpoint testing (use http-api-test).
effort: medium
---

Every Go test this skill writes uses the same table-driven shape. The point is uniformity: a reader
should be able to open any `_test.go` file in the repo and find the same `tests := []struct{...}`
table, the same `for _, tt := range tests` loop, and the same assertion calls. Do not invent a
variant shape because a case "doesn't need the table" — a single-case test still gets the table.

## The format

Non-negotiable rules. These are what make the tests uniform.

1. **One test function per function/method under test.** Name it `Test{FUNCTION_NAME}` for a plain
   function, `Test{STRUCT_NAME}_{METHOD_NAME}` for a method.
2. **Cases live in an anonymous struct slice named `tests`.** Each scenario is exactly one row.
3. **Arguments go in an `args` struct; struct dependencies go in a `fields` struct.** Never capture
   inputs from the enclosing scope.
4. **Iterate with `for _, tt := range tests` and run each row via `t.Run(tt.name, ...)`.** The
   subtest name is always `tt.name`.
5. **Errors are asserted with `assert.ErrorAssertionFunc`,** so each row decides its own expectation
   (`assert.NoError` / `assert.Error`). Omit the `wantErr` field entirely if the callee cannot
   return an error — never leave a declared field unused.
6. **Mocks are configured in `before`, released in `after`.** Both are optional per row; both must be
   nil-checked before being called. `after` is deferred so it runs even when an assertion fails.
7. **Interface-typed fields are declared as their mock type** in `fields`, not as the interface.

## Templates

### Plain function

```go
func Test{FUNCTION_NAME}(t *testing.T) {
	type args struct {
		// One field per parameter of {FUNCTION_NAME}.
	}

	tests := []struct {
		name    string
		args    args
		want    {RETURN_TYPE}
		wantErr assert.ErrorAssertionFunc // Omit if {FUNCTION_NAME} returns no error.
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := {FUNCTION_NAME}({ARGS})
			if !tt.wantErr(t, err, fmt.Sprintf("{FUNCTION_NAME}(%v)", {ARGS})) {
				return
			}
			assert.Equalf(t, tt.want, got, "{FUNCTION_NAME}(%v)", {ARGS})
		})
	}
}
```

If the function returns no error, drop the `wantErr` field and collapse the body to the single
`assert.Equalf` on `got`.

### Struct method

```go
func Test{STRUCT_NAME}_{METHOD_NAME}(t *testing.T) {
	type fields struct {
		// One field per dependency of {STRUCT_NAME}.
		// Interface-typed dependencies are declared as their MOCK type, e.g.
		//   repo *mocks.UserRepository
	}

	type args struct {
		// One field per parameter of {METHOD_NAME}, including ctx if present.
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    {RETURN_TYPE}
		wantErr assert.ErrorAssertionFunc // Omit if {METHOD_NAME} returns no error.
		before  func(fields, args)        // Arrange mocks. Optional.
		after   func(fields, args)        // Assert expectations / clean up. Optional.
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.before != nil {
				tt.before(tt.fields, tt.args)
			}
			if tt.after != nil {
				defer tt.after(tt.fields, tt.args)
			}

			u := &{STRUCT_NAME}{
				// Wire tt.fields into the struct.
			}

			got, err := u.{METHOD_NAME}({ARGS})
			if !tt.wantErr(t, err, fmt.Sprintf("{METHOD_NAME}(%v)", {ARGS})) {
				return
			}
			assert.Equalf(t, tt.want, got, "{METHOD_NAME}(%v)", {ARGS})
		})
	}
}
```

## Worked example

```go
func TestUserService_GetByID(t *testing.T) {
	type fields struct {
		repo *mocks.UserRepository
	}

	type args struct {
		ctx context.Context
		id  int64
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *User
		wantErr assert.ErrorAssertionFunc
		before  func(fields, args)
		after   func(fields, args)
	}{
		{
			name:   "returns the user when the repository finds it",
			fields: fields{repo: mocks.NewUserRepository(t)},
			args:   args{ctx: context.Background(), id: 1},
			want:   &User{ID: 1, Name: "Ada"},
			wantErr: assert.NoError,
			before: func(f fields, a args) {
				f.repo.On("FindByID", a.ctx, a.id).Return(&User{ID: 1, Name: "Ada"}, nil)
			},
			after: func(f fields, a args) {
				f.repo.AssertExpectations(t)
			},
		},
		{
			name:    "propagates the repository error",
			fields:  fields{repo: mocks.NewUserRepository(t)},
			args:    args{ctx: context.Background(), id: 2},
			want:    nil,
			wantErr: assert.Error,
			before: func(f fields, a args) {
				f.repo.On("FindByID", a.ctx, a.id).Return(nil, ErrNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.before != nil {
				tt.before(tt.fields, tt.args)
			}
			if tt.after != nil {
				defer tt.after(tt.fields, tt.args)
			}

			u := &UserService{repo: tt.fields.repo}

			got, err := u.GetByID(tt.args.ctx, tt.args.id)
			if !tt.wantErr(t, err, fmt.Sprintf("GetByID(%v, %v)", tt.args.ctx, tt.args.id)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetByID(%v, %v)", tt.args.ctx, tt.args.id)
		})
	}
}
```

Note the second case omits `after` — the field is nil and the nil-check skips it. That is the
intended way to opt out of a hook, not deleting the field from the table.

## Procedure

1. **Read the target.** Take the signature of the function/method and the scenarios to cover. One
   scenario becomes one row.
2. **Check what's already there.** If a test for the target exists and is not in this format, rewrite
   it into this format rather than appending alongside it. Reuse the mocking library the repo already
   uses (look for `mockery`, `gomock`, or hand-written fakes) — do not introduce a second one.
3. **Fill the table.** Add every scenario as a row before touching the loop body; the loop body is
   boilerplate and comes straight from the template.
4. **Run the tests.** `go test ./...` on the affected package. The tests must compile and pass (or
   fail for the reason the scenario intends) before you report done.
