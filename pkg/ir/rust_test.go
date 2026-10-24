package ir

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRustArguments(t *testing.T) {
	testcases := []struct {
		input   string
		want    []*TriggerArgument
		wantErr string
	}{
		{
			input:   "u32",
			wantErr: "name: Type",
		},
		{
			input:   "u32 foo",
			wantErr: "name: Type",
		},
		{
			input:   "foo:",
			wantErr: "has no type",
		},
		{
			input: "foo: u32",
			want: []*TriggerArgument{
				{Name: "foo", Type: "u32"},
			},
		},
		{
			input: "foo: u32,",
			want: []*TriggerArgument{
				{Name: "foo", Type: "u32"},
			},
		},
		{
			input:   "foo: u32, foo: u64",
			wantErr: "argument is defined twice",
		},
		{
			input: "job: crate::settlement::HaulJob",
			want: []*TriggerArgument{
				{Name: "job", Type: "crate::settlement::HaulJob"},
			},
		},
		{
			input: "map: HashMap<u32, Vec<(Fx,   Fx)>>, ids: [u8; 4], pair: (u8, u8)",
			want: []*TriggerArgument{
				{Name: "map", Type: "HashMap<u32, Vec<(Fx, Fx)>>"},
				{Name: "ids", Type: "[u8; 4]"},
				{Name: "pair", Type: "(u8, u8)"},
			},
		},
		{
			input:   "foo: &u32",
			wantErr: "references and lifetimes are not allowed",
		},
		{
			input:   "foo: Cow<'static, str>",
			wantErr: "references and lifetimes are not allowed",
		},
		{
			input:   "foo: Vec<u32",
			wantErr: "unterminated",
		},
		{
			input:   "foo: Vec<u32)>",
			wantErr: "unmatched",
		},
		{
			input:   "1foo: u32",
			wantErr: "invalid argument name",
		},
		{
			input:   "foo: *const u32",
			wantErr: "invalid char",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParseRustArguments(tc.input)
			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
