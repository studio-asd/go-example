package postgres

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
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
				RoleUUID:      uuid.New(),
				RoleName:      "simple_role",
				CreatedAt:     createdAt,
				PermissionIDs: []int64{1, 2, 3},
			},
			expect: RolePermissions{
				RoleID:   1,
				RoleName: "simple_role",
				Permissions: []Permission{
					{
						PermissionID:    1,
						PermissionName:  "one",
						PermissionType:  "API",
						PermissionKey:   "A",
						PermissionValue: "W",
					},
					{
						PermissionID:    2,
						PermissionName:  "two",
						PermissionType:  "API",
						PermissionKey:   "B",
						PermissionValue: "R",
					},
					{
						PermissionID:    3,
						PermissionName:  "three",
						PermissionType:  "API",
						PermissionKey:   "C",
						PermissionValue: "D",
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

	// Need to create the permission first for the initial setup.
	permIDs, err := tq.CreatePermissions(
		t.Context(),
		[]SecurityPermission{
			{
				PermissionUuid:  uuid.New(),
				PermissionName:  "one",
				PermissionType:  "API",
				PermissionKey:   "A",
				PermissionValue: "W",
			},
			{
				PermissionUuid:  uuid.New(),
				PermissionName:  "two",
				PermissionType:  "API",
				PermissionKey:   "B",
				PermissionValue: "R",
			},
			{
				PermissionUuid:  uuid.New(),
				PermissionName:  "three",
				PermissionType:  "API",
				PermissionKey:   "C",
				PermissionValue: "D",
			},
		}...,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(permIDs) != 3 {
		t.Fatal("expecting to create 3 permissions")
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			t.Log("schema_name", th.DefaultSearchPath())
			roleID, err := tq.CreateRole(t.Context(), test.create)
			if err != test.err {
				t.Fatalf("expecting error %v but got %v", test.err, err)
			}

			got, err := tq.GetRolePermissions(t.Context(), roleID)
			if err != nil {
				t.Fatal(err)
			}

			opts := []cmp.Option{
				cmpopts.IgnoreFields(
					Permission{}, "PermissionUUID", "CreatedAt", "UpdatedAt",
				),
			}
			if diff := cmp.Diff(test.expect, got, opts...); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}
