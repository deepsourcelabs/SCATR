package pragma

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewFile(t *testing.T) {
	type args struct {
		content       string
		commentPrefix []string
	}
	tests := []struct {
		name string
		args args
		want map[int]*Pragma
	}{
		{
			name: "no pragmas - go",
			args: args{
				content: `
package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello World") // something just like this
}
`,
				commentPrefix: []string{"//"},
			},
			want: map[int]*Pragma{},
		},
		{
			name: "no message or column - go",
			args: args{
				content: `
package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello World") // [GO-W1000]
}
`,
				commentPrefix: []string{"//"},
			},
			want: map[int]*Pragma{
				9: {
					Issues: map[string][]*Issue{"GO-W1000": {}},
					Hit:    map[string]bool{"GO-W1000": false},
				},
			},
		},
		{
			name: "mixed messages and columns - go",
			args: args{
				content: `
package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello World") // [GO-W1000]: 10 "Hello", 20, "World"
}
`,
				commentPrefix: []string{"//"},
			},
			want: map[int]*Pragma{
				9: {
					Issues: map[string][]*Issue{
						"GO-W1000": {
							{Message: "Hello", Column: 10},
							{Message: "", Column: 20},
							{Message: "World", Column: 0},
						},
					},
					Hit: map[string]bool{"GO-W1000": false},
				},
			},
		},
		{
			name: "mixed messages and columns on previous line - go",
			args: args{
				content: `
package main

import (
	"fmt"
)

func main() {
	// [GO-W1000]: 10 "Hello", 20, "World"
	fmt.Println("Hello World")
}
`,
				commentPrefix: []string{"//"},
			},
			want: map[int]*Pragma{
				10: {
					Issues: map[string][]*Issue{
						"GO-W1000": {
							{Message: "Hello", Column: 10},
							{Message: "", Column: 20},
							{Message: "World", Column: 0},
						},
					},
					Hit: map[string]bool{"GO-W1000": false},
				},
			},
		},
		{
			name: "no pragmas - python",
			args: args{
				content: `
print("Hello World") # something just like this
`,
				commentPrefix: []string{"#"},
			},
			want: map[int]*Pragma{},
		},
		{
			name: "no message or column - python",
			args: args{
				content: `
print("Hello World") # [PY-W1000]
`,
				commentPrefix: []string{"#"},
			},
			want: map[int]*Pragma{
				2: {
					Issues: map[string][]*Issue{"PY-W1000": {}},
					Hit:    map[string]bool{"PY-W1000": false},
				},
			},
		},
		{
			name: "mixed messages and columns - python",
			args: args{
				content: `
print("Hello World") # [PY-W1000]: 10 "Hello", 20, "World"
`,
				commentPrefix: []string{"#"},
			},
			want: map[int]*Pragma{
				2: {
					Issues: map[string][]*Issue{
						"PY-W1000": {
							{Message: "Hello", Column: 10},
							{Message: "", Column: 20},
							{Message: "World", Column: 0},
						},
					},
					Hit: map[string]bool{"PY-W1000": false},
				},
			},
		},
		{
			name: "mixed messages and columns on previous line - python",
			args: args{
				content: `
# [PY-W1000]: 10 "Hello", 20, "World"
print("Hello World")
`,
				commentPrefix: []string{"#"},
			},
			want: map[int]*Pragma{
				3: {
					Issues: map[string][]*Issue{
						"PY-W1000": {
							{Message: "Hello", Column: 10},
							{Message: "", Column: 20},
							{Message: "World", Column: 0},
						},
					},
					Hit: map[string]bool{"PY-W1000": false},
				},
			},
		},
		{
			name: "pragmas split on multiple lines",
			args: args{
				content: `
// [GO-W1000]: 10 "Hello", 20; [GO-W1001]
// [GO-W1002]: 30
fmt.Println("Hello")
`,
				commentPrefix: []string{"//"},
			},
			want: map[int]*Pragma{
				4: {
					Issues: map[string][]*Issue{
						"GO-W1000": {
							{Message: "Hello", Column: 10},
							{Message: "", Column: 20},
						},
						"GO-W1001": {},
						"GO-W1002": {
							{Message: "", Column: 30},
						},
					},
					Hit: map[string]bool{
						"GO-W1000": false,
						"GO-W1001": false,
						"GO-W1002": false,
					},
				},
			},
		},
		{
			name: "pragmas split on multiple lines - same line pragma",
			args: args{
				content: `
// [GO-W1000]: 10 "Hello", 20; [GO-W1001]
fmt.Println("Hello") // [GO-W1002]: 30
`,
				commentPrefix: []string{"//"},
			},
			want: map[int]*Pragma{
				3: {
					Issues: map[string][]*Issue{
						"GO-W1000": {
							{Message: "Hello", Column: 10},
							{Message: "", Column: 20},
						},
						"GO-W1001": {},
						"GO-W1002": {
							{Message: "", Column: 30},
						},
					},
					Hit: map[string]bool{
						"GO-W1000": false,
						"GO-W1001": false,
						"GO-W1002": false,
					},
				},
			},
		},
		{
			name: "pragmas split on multiple lines - edge case",
			args: args{
				content: `
// [GO-W1000]: 10 "Hello", 20; [GO-W1001]
fmt.Println("Hello") // [GO-W1002]: 30
fmt.Println("Hello") // [GO-W1003]: 30
`,
				commentPrefix: []string{"//"},
			},
			want: map[int]*Pragma{
				3: {
					Issues: map[string][]*Issue{
						"GO-W1000": {
							{Message: "Hello", Column: 10},
							{Message: "", Column: 20},
						},
						"GO-W1001": {},
						"GO-W1002": {
							{Message: "", Column: 30},
						},
					},
					Hit: map[string]bool{
						"GO-W1000": false,
						"GO-W1001": false,
						"GO-W1002": false,
					},
				},
				4: {
					Issues: map[string][]*Issue{
						"GO-W1003": {
							{Message: "", Column: 30},
						},
					},
					Hit: map[string]bool{
						"GO-W1003": false,
					},
				},
			},
		},
		{
			name: "multiple prefixes with mixed messages and columns with multiple messages - vue",
			args: args{
				content: `
<template>
	<h1>Hello Vue</h1>
	<!-- [VUE-W1000]: 10 "Hello", 20, "World"; [VUE-W1002]: 20 "Vue" -->
</template>
<script>
	// [JS-W1000]: 10 "Hello", 20; [JS-W1002]: "Something"; [JS-W1003]
</script>
`,
				commentPrefix: []string{"<!--", "//"},
			},
			want: map[int]*Pragma{
				5: {
					Issues: map[string][]*Issue{
						"VUE-W1000": {
							{Message: "Hello", Column: 10},
							{Message: "", Column: 20},
							{Message: "World", Column: 0},
						},
						"VUE-W1002": {
							{Message: "Vue", Column: 20},
						},
					},
					Hit: map[string]bool{
						"VUE-W1000": false,
						"VUE-W1002": false,
					},
				},
				8: {
					Issues: map[string][]*Issue{
						"JS-W1000": {
							{Message: "Hello", Column: 10},
							{Message: "", Column: 20},
						},
						"JS-W1002": {
							{Message: "Something", Column: 0},
						},
						"JS-W1003": {},
					},
					Hit: map[string]bool{
						"JS-W1000": false,
						"JS-W1002": false,
						"JS-W1003": false,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFile(tt.args.content, tt.args.commentPrefix); !reflect.DeepEqual(got.Pragmas, tt.want) {
				t.Errorf("NewFile() = %v, want %v, diff %v", got.Pragmas, tt.want,
					cmp.Diff(tt.want, got.Pragmas))
			}
		})
	}
}
