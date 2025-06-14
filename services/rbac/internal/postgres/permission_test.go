package postgres

import "testing"

// permKeysStub is a stub data to create the permission keys.
var permKeysStub = map[string]string{
	"user":   "API",
	"ledger": "API",
	"wallet": "API",
}

func TestCreatePermissionKeys(t *testing.T) {
}
