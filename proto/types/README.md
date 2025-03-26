# Types Proto

The `types` protobuf defines messages or data structure that can be shared across versions of `api`. The `types` are not bounded by the `api` version since it directly related to the internal data structure of the program. Different `api` version can use different versions of `types` depending the underlying data structure used by the `api`.
