# Errors

The errors package provides a way for the user to use a custom error type to insert more context/information about
the error.

## Custom Properties

### Kind

Kind is an identifier for both client and server of what `kind` of error is occured at the moment. As this package inspired by
upspin's [error package](https://github.com/upspin/upspin), we also adopt [kind](https://github.com/upspin/upspin/blob/master/errors/errors.go#L73)
from there.

As of now, there are several type of `kind`:

1. Bad Request
2. Unauthorized
3. Internal Error

### Fields

Fields is a key-value of `[]any`. This type is usually used to put more context to the error for `logging` purpose. This is also why
it has a method to convert the type to `slog.Attributes`.
