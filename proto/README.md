# Protobuf

This directory contains all the Types, APIs and protobuf files.

## Design

By design, there are two big folders that you need to care about.

1. [Types](./types/README.md)

    Types stores all the core types and underlying data structures for the program.

2. [API](./api/)

    API stores all API definition to be consumed by the client.

### Why Defining Internal Data Structures In Protobuf Types?

Sometimes, the API need to access the data structures/types directly and it is kinda annoying to cast the types back and front from the internal to the protobuf data types. Moreover, it can be confusing which API versions are compatible with which version of data structures? By importing the `types` directly from `api`, we can reduce the confusion.

So, this mean `v1`, `v2`, and `v3` APIs can use the `v1` type? Absolutely, the API surface usually evolves rather quickly than the internal data structures. This is because the internal data structures have been used by the current client and ensuring nothing breaks while migrating to the new one is not trivial, so it usually takes more time if really needed.

> Some big projects usually does this, but this might be overkill for starter as there are too many separations and things that we need to take care about. The structure is intended to create a showcase of when this kind of thing is used in a project.

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
