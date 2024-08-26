package env

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		envs     []string
		keys     []string
		defaults []string
		expects  []string
	}{
		{
			name: "all exist",
			envs: []string{
				"key1=value1",
				"key2=value2",
			},
			keys: []string{
				"key1",
				"key2",
			},
			defaults: []string{
				"",
				"",
			},
			expects: []string{
				"value1",
				"value2",
			},
		},
		{
			name: "some not exist",
			envs: []string{
				"key1=value1",
				"key2=value2",
			},
			keys: []string{
				"key1",
				"key2",
				"key3",
				"key4",
			},
			defaults: []string{
				"",
				"",
				"three",
				"four",
			},
			expects: []string{
				"value1",
				"value2",
				"three",
				"four",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, env := range test.envs {
				kv := strings.Split(env, "=")
				t.Setenv(kv[0], kv[1])
			}
			var got []string
			for idx, k := range test.keys {
				g := GetEnvOrDefault(k, test.defaults[idx])
				got = append(got, g)
			}

			if diff := cmp.Diff(test.expects, got); diff != "" {
				t.Fatalf("(-want/+got)\n%s", diff)
			}
		})
	}
}
