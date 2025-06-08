package permission

import "testing"

func TestHas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		perms  Permission
		check  Permission
		expect bool
	}{
		{
			perms:  Read,
			check:  Read,
			expect: true,
		},
		{
			perms:  Read,
			check:  Write,
			expect: false,
		},
		{
			perms:  Read,
			check:  Delete,
			expect: false,
		},
		{
			perms:  Write,
			check:  Write,
			expect: true,
		},
		{
			perms:  Write,
			check:  Read,
			expect: false,
		},
		{
			perms:  Write,
			check:  Delete,
			expect: false,
		},
		{
			perms:  Delete,
			check:  Delete,
			expect: true,
		},
		{
			perms:  Delete,
			check:  Read,
			expect: false,
		},
		{
			perms:  Delete,
			check:  Write,
			expect: false,
		},
		{
			perms:  Read | Write | Delete,
			check:  Read,
			expect: true,
		},
		{
			perms:  Read | Write | Delete,
			check:  Write,
			expect: true,
		},
		{
			perms:  Read | Write | Delete,
			check:  Delete,
			expect: true,
		},
	}

	for _, test := range tests {
		t.Run(test.perms.String()+"_has_"+test.check.String(), func(t *testing.T) {
			got := test.perms.Has(test.check)
			if test.expect != got {
				t.Fatalf("expecting %v but got %v", test.expect, got)
			}
		})
	}
}
