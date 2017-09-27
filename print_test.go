package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestPrintJSON(t *testing.T) {
	type someType struct {
		Name string
		Age  int
	}

	foo := someType{
		Name: "Test",
		Age:  42,
	}

	out := new(bytes.Buffer)
	require.NoError(t, PrintWriter("json", foo, out))
	assert.NotEmpty(t, out)

	t.Log("\n" + out.String())

	var bar someType
	require.NoError(t, json.Unmarshal(out.Bytes(), &bar))
	assert.Equal(t, foo, bar)
}

func TestPrintYAML(t *testing.T) {
	type someType struct {
		Name string
		Age  int
	}

	foo := someType{
		Name: "Test",
		Age:  42,
	}

	out := new(bytes.Buffer)
	require.NoError(t, PrintWriter("yaml", foo, out))
	assert.NotEmpty(t, out)

	t.Log("\n" + out.String())

	var bar someType
	require.NoError(t, yaml.Unmarshal(out.Bytes(), &bar))
	assert.Equal(t, foo, bar)
}

func TestPrintTable(t *testing.T) {
	cases := map[string]struct {
		instance interface{}
		expected []string
	}{
		"no tags": {
			instance: struct {
				Name  string
				Age   int
				Value bool
			}{
				Name:  "Test",
				Age:   42,
				Value: true,
			},
			expected: []string{
				"NAME    AGE     VALUE",
				"Test    42      true    ",
			},
		},
		"with ignore tags": {
			instance: struct {
				Name  string
				Age   int
				Value bool `table:"-"`
			}{
				Name:  "Test",
				Age:   42,
				Value: true,
			},
			expected: []string{
				"NAME    AGE",
				"Test    42      ",
			},
		},
		"rename columns": {
			instance: struct {
				Name  string `table:"key"`
				Age   int    `table:"age"`
				Value bool   `table:"-"`
			}{
				Name:  "Test",
				Age:   42,
				Value: true,
			},
			expected: []string{
				"key     age",
				"Test    42      ",
			},
		},
		"slice": {
			instance: []struct {
				Name  string
				Age   int
				Value bool
			}{
				{Name: "Foo", Age: 1, Value: true},
				{Name: "Bar", Age: 2, Value: false},
				{Name: "Baz", Age: 3, Value: false},
			},
			expected: []string{
				"NAME    AGE     VALUE",
				"Foo     1       true    ",
				"Bar     2       false   ",
				"Baz     3       false   ",
			},
		},
		"slice of strings": {
			instance: []string{"A", "B", "C"},
			expected: []string{
				"A",
				"B",
				"C",
			},
		},
		"slice of ints": {
			instance: []int{1, 2, 3},
			expected: []string{
				"1",
				"2",
				"3",
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			out := new(bytes.Buffer)
			require.NoError(t, PrintWriter("table", c.instance, out))
			assert.Equal(t, strings.Join(c.expected, "\n")+"\n", out.String())
		})
	}
}
