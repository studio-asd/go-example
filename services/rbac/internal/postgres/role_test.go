package postgres

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestCreateRole(t *testing.T) {
	t.Parallel()

	createdAt := time.Now()
	tests := []struct {
		name   string
		create CreateRole
		expect RolePermissions
		err    error
	}{
		{
			name: "create a simple role",
			create: CreateRole{
				RoleExternalID: "one",
				RoleName:       "simple_role",
				CreatedAt:      createdAt,
				PermissionIDs:  []int64{1, 2, 3},
			},
			expect: RolePermissions{
				RoleID:   1,
				RoleName: "simple_role",
				Permissions: []Permission{
					{
						PermissionID: 1,
					},
					{
						PermissionID: 2,
					},
					{
						PermissionID: 3,
					},
				},
			},
			err: nil,
		},
	}

	th, err := testHelper.ForkPostgresSchema(t.Context(), testHelper.Postgres(), "public")
	if err != nil {
		t.Fatal(err)
	}
	tq := New(th.Postgres())

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			roleID, err := tq.CreateRole(t.Context(), test.create)
			if err != test.err {
				t.Fatalf("expecting error %v but got %v", test.err, err)
			}

			got, err := tq.GetRolePermissions(t.Context(), roleID)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.expect, got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}
