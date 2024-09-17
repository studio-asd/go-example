# Protobuf

This directory contains all the APIs and protobuf files.

## Protovalidate

We use [protovalidate](https://github.com/bufbuild/protovalidate) to validate our proto messages.

As it uses a specific [contraint_id]() to define the error, we have the list of constraints and ids reserved for this.
Because the ids are reserved, validations that falls into the expression of constraint **must** use the same id.

Below is the list `constraint_id` and example of the validation.

### Email Validation

ID: `validate.email`

```proto3
message EmailValidation {
  string email = 1 [(buf.validate.field).cel = {
    id: "validate.email"
    message: "must be a valid email"
    expression: "this.isEmail()"
  }];
}
```

Buf's example: [click link](https://github.com/bufbuild/protovalidate/blob/main/examples/cel_string_is_email.proto)

### IP Validation

ID: `validate.ip`

```proto3
message IPValidation {
  string email = 1 [(buf.validate.field).cel = {
    id: "validate.ip"
    message: "must be a valid ip"
    expression: "this.isIp()"
  }];
}
```

Buf's example: [click link](https://github.com/bufbuild/protovalidate/blob/main/examples/cel_string_is_ip.proto)
