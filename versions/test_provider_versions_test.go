package versions

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSemVer_Less(t *testing.T) {
	tests := []struct {
		name  string
		this  string
		other string
		want  bool
	}{
		{
			"simple", "0", "0", false,
		},
		{
			"full_equals", "1.0.0", "1.0.0", false,
		},
		{
			"shorter_length", "1.0", "1.0.1", true,
		},
		{
			"shorter_length", "1.0", "1.0.1", true,
		},
		{
			"major_less", "1.1.0", "2.1.0", true,
		},
		{
			"minor_less", "1.1.0", "1.2.0", true,
		},
		{
			"patch_less", "1.1.0", "1.1.1", true,
		},
		{
			"major_less_inequal_length", "1.1.0", "2", true,
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			this := MakeSemVerFromString(tt.this)
			other := MakeSemVerFromString(tt.other)
			result := this.Less(other)
			op_result := other.Less(this)
			if result != tt.want {
				if tt.want {
					t.Errorf("%s is not less than %s", tt.this, tt.other)
				} else {
					t.Errorf("%s is less than %s", tt.this, tt.other)
				}
			}
			if tt.this != tt.other && op_result == result {
				t.Errorf("inverse %s.Less(%s) incorrect", tt.other, tt.this)
			}
		})
	}
}

func TestProviderVersions_Merge(t *testing.T) {
	tests := []struct {
		name   string
		target string
		source string
		want   string
	}{
		{"add_new_version",
			`{"versions":[{"version":"0.6.1","protocols":["5.0"],"platforms":[{"os":"darwin","arch":"amd64"}]}]}`,
			`{"versions":[{"version":"0.6.2","protocols":["5.0"],"platforms":[{"os":"darwin","arch":"amd64"}]}]}`,
			`{"versions":[{"version":"0.6.1","protocols":["5.0"],"platforms":[{"os":"darwin","arch":"amd64"}]}, {"version":"0.6.2","protocols":["5.0"],"platforms":[{"os":"darwin","arch":"amd64"}]}]}`,
		},
		{"add_previous_version",
			`{"versions":[{"version":"0.6.1","protocols":["5.0"],"platforms":[{"os":"darwin","arch":"amd64"}]}]}`,
			`{"versions":[{"version":"0.5.9","protocols":["5.0"],"platforms":[{"os":"darwin","arch":"amd64"}]}]}`,
			`{"versions":[{"version":"0.5.9","protocols":["5.0"],"platforms":[{"os":"darwin","arch":"amd64"}]}, {"version":"0.6.1","protocols":["5.0"],"platforms":[{"os":"darwin","arch":"amd64"}]}]}`,
		},
		{"add_a_new_protocol",
			`{"versions":[{"version":"0.6.1","protocols":["5.1"],"platforms":[{"os":"darwin","arch":"amd64"}]}]}`,
			`{"versions":[{"version":"0.6.1","protocols":["5.0"],"platforms":[{"os":"darwin","arch":"amd64"}]}]}`,
			`{"versions":[{"version":"0.6.1","protocols":["5.0", "5.1"], "platforms":[{"os":"darwin","arch":"amd64"}]}]}`,
		},
		{"add_a_protocol_and_a_platform",
			`{"versions":[{"version":"0.6.1","protocols":["5.0"],"platforms":[{"os":"darwin","arch":"amd64"}]}]}`,
			`{"versions":[{"version":"0.6.1","protocols":["5.1"],"platforms":[{"os":"darwin","arch":"amd64"}, {"os":"linux","arch":"amd64"}]}]}`,
			`{"versions":[{"version":"0.6.1","protocols":["5.0", "5.1"], "platforms":[{"os":"darwin","arch":"amd64"}, {"os":"linux","arch":"amd64"}]}]}`,
		},
		{"sort_added_a_protocol_and_a_platform",
			`{"versions":[{"version":"0.6.1","protocols":["5.1"],"platforms":[{"os":"linux","arch":"amd64"}]}]}`,
			`{"versions":[{"version":"0.6.1","protocols":["5.0"],"platforms":[{"os":"darwin","arch":"amd64"}]}]}`,
			`{"versions":[{"version":"0.6.1","protocols":["5.0", "5.1"], "platforms":[{"os":"darwin","arch":"amd64"}, {"os":"linux","arch":"amd64"}]}]}`,
		},
		{"empty_target",
			`{}`,
			`{"versions":[{"version":"0.6.2","protocols":["5.0"],"platforms":[{"os":"darwin","arch":"amd64"}]}]}`,
			`{"versions":[{"version":"0.6.2","protocols":["5.0"],"platforms":[{"os":"darwin","arch":"amd64"}]}]}`,
		},
		{"empty_source",
			`{"versions":[{"version":"0.6.2","protocols":["5.0"],"platforms":[{"os":"darwin","arch":"amd64"}]}]}`,
			`{}`,
			`{"versions":[{"version":"0.6.2","protocols":["5.0"],"platforms":[{"os":"darwin","arch":"amd64"}]}]}`,
		},
		{"empty_source_and_target",
			`{}`,
			`{}`,
			`{}`,
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			var target ProviderVersions
			var source ProviderVersions
			var want ProviderVersions
			if err := json.Unmarshal([]byte(tt.target), &target); err != nil {
				t.Fatalf("failed to unmarshal merge")
			}
			if err := json.Unmarshal([]byte(tt.source), &source); err != nil {
				t.Fatalf("failed to unmarshal source")
			}
			if err := json.Unmarshal([]byte(tt.want), &want); err != nil {
				t.Fatalf("failed to unmarshal want")
			}
			target.Merge(source)
			if !reflect.DeepEqual(target, want) {
				t.Errorf("merge did not deliver desired result")
			}
		})
	}
}
