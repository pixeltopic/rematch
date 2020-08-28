package rematch

import (
	"errors"
	"reflect"
	"testing"
)

func Test_tokenizeExpr(t *testing.T) {
	type args struct {
		expr string
	}
	tests := []struct {
		name    string
		args    args
		want    []token
		wantErr bool
		err     error
	}{
		{
			name: "expr with quoted unescaped wildcards should pass",
			args: args{expr: "\"**\""},
			want: []token{{Str: "\"**\""}},
		},
		{
			name: "expr with escaped backslash should pass",
			args: args{expr: "\"\\\\\""},
			want: []token{{Str: "\"\\\\\""}},
		},
		{
			name: "expr with escaped backslash and regex should pass",
			args: args{expr: "\"\\\\\\?\""},
			want: []token{{Str: "\"\\\\\\?\"", Regex: true}},
		},
		{
			name: "expr with invalid adjacent operands will not fail and be caught during shunting",
			args: args{expr: "\"foo\"\"wild\\*card\"\"bar\""},
			want: []token{{Str: "\"foo\""}, {Str: "\"wild\\*card\"", Regex: true}, {Str: "\"bar\""}},
		},
		/* Error tests */
		{
			name:    "expr with only escaped wildcards should not pass (1)",
			args:    args{expr: "!\"\\*\\*\\*\\?\""},
			wantErr: true,
			err:     errOnlyWildcards,
		},
		{
			name:    "expr with only escaped wildcards should not pass (2)",
			args:    args{expr: "!\"\\_\\*\\*\\_\""},
			wantErr: true,
			err:     errOnlyWildcards,
		},
		{
			name:    "expr with mismatched quotations should not pass (1)",
			args:    args{expr: "\""},
			wantErr: true,
			err:     errMismatchedQuotations,
		},
		{
			name:    "expr with mismatched quotations should not pass (2)",
			args:    args{expr: "\"foo"},
			wantErr: true,
			err:     errMismatchedQuotations,
		},
		{
			name:    "expr with mismatched quotations should not pass because it is considered unquoted",
			args:    args{expr: "ddd\""},
			wantErr: true,
			err:     errInvalidChar,
		},
		{
			name:    "expr with an empty quoted word should not pass",
			args:    args{expr: "\"\"+foo"},
			wantErr: true,
			err:     SyntaxError("invalid word; no quoted pattern"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tokenizeExpr(tt.args.expr)
			if err != nil {
				if tt.wantErr && !errors.Is(err, tt.err) {
					t.Errorf("tokenizeExpr() error=%v, expected error=%v", err, tt.err)
					return
				}
				if !tt.wantErr {
					t.Errorf("tokenizeExpr() error=%v, wantErr=%v", err, tt.wantErr)
					return
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tokenizeExpr() got=%v, want=%v", got, tt.want)
			}
		})
	}
}
