package main

import (
	"strings"
	"testing"
)

func runGPM(t *testing.T, in string) string {
	t.Helper()
	g := NewGPM()
	v, err := g.Expand(in)
	if err != nil {
		t.Fatalf("expand error: %v", err)
	}
	return v
}

func TestExpansionRegressions(t *testing.T) {
	tests := []struct{ name, input, want string }{
		{"plain whitespace", "hello world\nnext line\t \n", "hello world\nnext line\t \n"},
		{"macro body whitespace", "define(MSG,hello world)MSG()\n", "hello world\n"},
		{"unknown call", "ordinary \t(a, b)", "ordinary \t(a, b)"},
		{"unknown unicode call", "ordinary(café, 日本語, 🙂)", "ordinary(café, 日本語, 🙂)"},
		{"unknown nested call", "define(ID,$1)ordinary ( ID(x), y )", "ordinary ( ID(x), y )"},
		{"unicode argument", "define(ID,$1)ID(café 日本語 🙂)", "café 日本語 🙂"},
		{"unicode identifier and whitespace", "define(挨拶,$1)挨拶\u2003(こんにちは)", "こんにちは"},
		{"quoted unicode", `define(ID,$1)ID("café, 日本語 (🙂)")`, `"café, 日本語 (🙂)"`},
		{"escaped quote", `define(ID,$1)ID("café\" , 🙂")`, `"café\" , 🙂"`},
		{"malformed call", "ordinary (café, x", "ordinary (café, x"},
		{"zero arguments", "define(COUNT,$#)COUNT() COUNT(  )", "0 0"},
		{"explicit empty arguments", `define(COUNT,$#)COUNT(,) COUNT(x,) COUNT("")`, "2 2 1"},
		{"expanded false condition", "define(FALSE,0)ifelse(FALSE(),yes,no)", "no"},
		{"expanded empty condition", "define(EMPTY,)ifelse(EMPTY(),yes,no)", "no"},
		{"expanded true condition", "define(TRUE,1)ifelse(TRUE(),yes,no)", "yes"},
		{"lazy branches", "define(LOOP,LOOP())ifelse(0,LOOP(),safe) ifelse(1,safe,LOOP())", "safe safe"},
		{"unselected side effect", "ifelse(0,define(X,bad),ok)X()", "okX()"},
		{"generated call", "define(NAME,ID)define(ID,$1)NAME()(hello)", "hello"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := runGPM(t, tc.input); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRecursionLimit(t *testing.T) {
	for _, input := range []string{
		"define(LOOP,LOOP())LOOP()",
		"define(A,B())define(B,A())A()",
		"define(LOOP,ifelse(1,LOOP(),done))LOOP()",
		"define(LOOP,ifelse(LOOP(),yes,no))LOOP()",
	} {
		t.Run(input, func(t *testing.T) {
			g := NewGPM()
			g.maxDepth = 8
			if _, err := g.Expand(input); err == nil || !strings.Contains(err.Error(), "max recursion depth") {
				t.Fatalf("expected recursion limit error, got %v", err)
			}
			if got, err := g.Expand("still usable"); err != nil || got != "still usable" {
				t.Fatalf("expansion after error: got %q, %v", got, err)
			}
		})
	}
}

func TestDepthBudgetAllowsIndependentCalls(t *testing.T) {
	g := NewGPM()
	g.maxDepth = 4
	g.Define("A", "B()")
	g.Define("B", "ok")
	input := strings.Repeat("A() ", 20)
	want := strings.Repeat("ok ", 20)
	if got, err := g.Expand(input); err != nil || got != want {
		t.Fatalf("got %q, %v; want %q", got, err, want)
	}
}

func TestExamples(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "greet",
			in:   `define(GREET, "hello, $1!")GREET(GPM)`,
			want: `"hello, GPM!"`,
		},
		{
			name: "arithmetic templating with nesting",
			in:   `define(ADD, "($1)+($2)")define(MUL, "($1)*($2)")ADD(2, MUL(3, 4))`,
			want: `"(2)+("(3)*(4)")"`,
		},
		{
			name: "dollar star and count",
			in:   `define(INFO, "args=$#, all=[$*]")INFO(a, b, c)`,
			want: `"args=3, all=[a,b,c]"`,
		},
		{
			name: "conditionals ifelse",
			in:   `define(CHOICE, "ifelse($1, 'YES', 'NO')")CHOICE(1)CHOICE(0)CHOICE(   )`,
			want: `"'YES'""'NO'""'NO'"`,
		},
		{
			name: "macro generates another macro",
			in:   `define(MAKE_PAIR, "define(PAIR, '($1,$2)')")MAKE_PAIR(x, y)PAIR(a, b)`,
			want: `""'(x,y)'`,
		},
		{
			name: "undef macro",
			in:   `define(VERSION, "1.2.3")VERSION()undef(VERSION)VERSION()`,
			want: `"1.2.3"VERSION()`,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := runGPM(t, tc.in)
			if got != tc.want {
				t.Fatalf("unexpected output.\n--- got ---\n%q\n--- want ---\n%q", got, tc.want)
			}
		})
	}
}
