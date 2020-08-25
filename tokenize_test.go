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
			name:    "expr with escaped backslash and regex should pass",
			args:    args{expr: "!\"\\*\\*\\*\\?\""},
			wantErr: true,
			err:     SyntaxError("invalid word; cannot only contain wildcards"),
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
