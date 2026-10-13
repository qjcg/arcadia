package desugar_test

import (
	"testing"

	"github.com/qjcg/arcadia/exp/terebra/internal/desugar"
)

// TestSource covers the Mechanism A contract from ADR 0004: unquoted
// auger pipes are rewritten into the reserved auger stage, and every
// other source is left byte-identical.
func TestSource(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "should_leave_plain_command_byte_identical",
			in:   "echo hi",
			want: "echo hi",
		},
		{
			name: "should_leave_or_operator_byte_identical",
			in:   "a || b",
			want: "a || b",
		},
		{
			name: "should_leave_pipe_both_operator_byte_identical",
			in:   "a |& b",
			want: "a |& b",
		},
		{
			name: "should_leave_double_pipe_then_redirect_byte_identical",
			in:   "a ||> b",
			want: "a ||> b",
		},
		{
			name: "should_rewrite_auger_pipe_into_reserved_stage",
			in:   "cmd1 |> cmd2",
			want: "cmd1 | __terebra_auger__ cmd2",
		},
		{
			name: "should_rewrite_glued_auger_pipe_into_reserved_stage",
			in:   "cmd1 |>cmd2",
			want: "cmd1 | __terebra_auger__ cmd2",
		},
		{
			name: "should_rewrite_glued_encoder_form_into_enc_flag",
			in:   "cmd1 |>json",
			want: "cmd1 | __terebra_auger__ --enc=json",
		},
		{
			name: "should_rewrite_spaced_encoder_form_into_enc_flag",
			in:   "cmd1 |> json",
			want: "cmd1 | __terebra_auger__ --enc=json",
		},
		{
			name: "should_rewrite_yaml_encoder_form_into_enc_flag",
			in:   "cmd |>yaml",
			want: "cmd | __terebra_auger__ --enc=yaml",
		},
		{
			name: "should_rewrite_cue_encoder_form_into_enc_flag",
			in:   "cmd |> cue",
			want: "cmd | __terebra_auger__ --enc=cue",
		},
		{
			name: "should_treat_encoder_keyword_with_args_as_stage_command",
			in:   "drill fs . -r |> cue vet ./schema.cue",
			want: "drill fs . -r | __terebra_auger__ cue vet ./schema.cue",
		},
		{
			name: "should_keep_encoder_form_before_redirect",
			in:   "cmd |> json > out",
			want: "cmd | __terebra_auger__ --enc=json > out",
		},
		{
			name: "should_leave_auger_pipe_in_single_quotes_byte_identical",
			in:   "echo '|>'",
			want: "echo '|>'",
		},
		{
			name: "should_leave_auger_pipe_in_double_quotes_byte_identical",
			in:   `echo "a |> b"`,
			want: `echo "a |> b"`,
		},
		{
			name: "should_leave_escaped_pipe_byte_identical",
			in:   `echo \|> x`,
			want: `echo \|> x`,
		},
		{
			name: "should_rewrite_auger_pipe_inside_command_substitution",
			in:   "echo $(a |> b)",
			want: "echo $(a | __terebra_auger__ b)",
		},
		{
			name: "should_rewrite_auger_pipe_inside_quoted_command_substitution",
			in:   `echo "x $(a |> b)"`,
			want: `echo "x $(a | __terebra_auger__ b)"`,
		},
		{
			name: "should_leave_quoted_auger_pipe_inside_substitution",
			in:   "echo $(echo '|>')",
			want: "echo $(echo '|>')",
		},
		{
			name: "should_rewrite_auger_pipe_inside_backticks",
			in:   "echo `a |> b`",
			want: "echo `a | __terebra_auger__ b`",
		},
		{
			name: "should_leave_auger_pipe_in_comment_byte_identical",
			in:   "echo a |> b # |> c",
			want: "echo a | __terebra_auger__ b # |> c",
		},
		{
			name: "should_leave_auger_pipe_in_heredoc_body_byte_identical",
			in:   "cat <<EOF |> wc\nx |> y\nEOF\n",
			want: "cat <<EOF | __terebra_auger__ wc\nx |> y\nEOF\n",
		},
		{
			name: "should_leave_auger_pipe_in_indented_heredoc_body",
			in:   "cat <<-EOF |> wc\n\tx |> y\n\tEOF\n",
			want: "cat <<-EOF | __terebra_auger__ wc\n\tx |> y\n\tEOF\n",
		},
		{
			name: "should_leave_auger_pipe_in_quoted_delimiter_heredoc",
			in:   "cat <<'EOF'\n|> x\nEOF\n",
			want: "cat <<'EOF'\n|> x\nEOF\n",
		},
		{
			name: "should_leave_auger_pipe_inside_arithmetic_expansion",
			in:   "echo $((x |> y))",
			want: "echo $((x |> y))",
		},
		{
			name: "should_rewrite_auger_pipe_with_background_operator",
			in:   "cmd1 |> cmd2 &",
			want: "cmd1 | __terebra_auger__ cmd2 &",
		},
		{
			name: "should_rewrite_only_auger_pipe_in_mixed_pipeline",
			in:   "a | b |> c |& d",
			want: "a | b | __terebra_auger__ c |& d",
		},
		{
			name: "should_rewrite_auger_pipe_following_parameter_expansion",
			in:   "echo ${x}|> y",
			want: "echo ${x}| __terebra_auger__ y",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := desugar.Source(tc.in)
			if got != tc.want {
				t.Errorf("Source(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
