package ir

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCppArguments(t *testing.T) {
	testcases := []struct {
		input   string
		want    []*TriggerArgument
		wantErr string
	}{
		{
			input:   "",
			wantErr: "at least 2 words",
		},
		{
			input:   "int",
			wantErr: "at least 2 words",
		},
		{
			input:   "int foo,",
			wantErr: "at least 2 words",
		},
		{
			input:   "int foo, std::string",
			wantErr: "at least 2 words",
		},
		{
			input: "int foo",
			want: []*TriggerArgument{
				{Name: "foo", Type: "int"},
			},
		},
		{
			input:   "int foo, int foo",
			wantErr: "argument is defined twice",
		},
		{
			input: "std::string foo",
			want: []*TriggerArgument{
				{Name: "foo", Type: "std::string"},
			},
		},
		{
			input: "const std::string& foo",
			want: []*TriggerArgument{
				{Name: "foo", Type: "const std::string&"},
			},
		},
		{
			input: "const std::vector<std::string>& foo",
			want: []*TriggerArgument{
				{Name: "foo", Type: "const std::vector<std::string>&"},
			},
		},
		{
			input: "const std::vector<std::unique_ptr<std::string, 22>>& foo",
			want: []*TriggerArgument{
				{Name: "foo", Type: "const std::vector<std::unique_ptr<std::string, 22>>&"},
			},
		},
		{
			input: "const std::vector<std::unique_ptr<std::string, 22>>& foo, int bar, std::string<22, std::greater()> baz",
			want: []*TriggerArgument{
				{Name: "foo", Type: "const std::vector<std::unique_ptr<std::string, 22>>&"},
				{Name: "bar", Type: "int"},
				{Name: "baz", Type: "std::string<22, std::greater()>"},
			},
		},
		{
			input:   "std::vector<int foo",
			wantErr: "unterminated template",
		},
		{
			input:   "std::vector<std::string>> foo",
			wantErr: "closing unopening template",
		},
		{
			input:   "int foo&",
			wantErr: "unexpected character in argument name",
		},
	}

	for _, tc := range testcases {
		got, gotErr := ParseCppArguments(tc.input)

		if tc.wantErr != "" {
			if gotErr == nil {
				t.Fatalf("Expected error, got none (value %v)", got)
				continue
			}

			assert.Contains(t, gotErr.Error(), tc.wantErr)
		} else if gotErr != nil {
			t.Fatalf("unexpected error: %q", gotErr)
		}

		assert.Equal(t, tc.want, got)
	}
}
