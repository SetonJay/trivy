package report_test

import (
	"strings"
	"testing"

	"github.com/aquasecurity/trivy/pkg/report"

	ftypes "github.com/aquasecurity/trivy/pkg/fanal/types"

	"github.com/stretchr/testify/assert"
)

func TestSecretRenderer(t *testing.T) {

	tests := []struct {
		name  string
		input []ftypes.SecretFinding
		want  string
	}{
		{
			name:  "no results",
			input: nil,
			want:  "",
		},
		{
			name: "single line",
			input: []ftypes.SecretFinding{
				{
					RuleID:    "rule-id",
					Category:  ftypes.SecretRuleCategory("category"),
					Title:     "this is a title",
					Severity:  "HIGH",
					StartLine: 1,
					EndLine:   1,
					Code: ftypes.Code{
						Lines: []ftypes.Line{
							{
								Number:     1,
								Content:    "password=secret",
								IsCause:    true,
								FirstCause: true,
								LastCause:  true,
							},
						},
					},
					Match: "secret",
				},
			},
			want: `HIGH: category (rule-id)
════════════════════════════════════════
this is a title
────────────────────────────────────────
 my-file:1
────────────────────────────────────────
   1 [ password=secret
────────────────────────────────────────


`,
		},
		{
			name: "multiple line",
			input: []ftypes.SecretFinding{
				{
					RuleID:    "rule-id",
					Category:  ftypes.SecretRuleCategory("category"),
					Title:     "this is a title",
					Severity:  "HIGH",
					StartLine: 3,
					EndLine:   4,
					Code: ftypes.Code{
						Lines: []ftypes.Line{
							{
								Number:  1,
								Content: "#!/bin/bash",
							},
							{
								Number:  2,
								Content: "",
							},
							{
								Number:     3,
								Content:    "password=this is a \\",
								IsCause:    true,
								FirstCause: true,
							},
							{
								Number:    4,
								Content:   "secret password",
								IsCause:   true,
								LastCause: true,
							},
							{
								Number:  5,
								Content: "some-app --password $password",
							},
							{
								Number:  6,
								Content: "echo all done",
							},
						},
					},
					Match: "secret",
				},
			},
			want: `HIGH: category (rule-id)
════════════════════════════════════════
this is a title
────────────────────────────────────────
 my-file:3-4
────────────────────────────────────────
   1   #!/bin/bash
   2   
   3 ┌ password=this is a \
   4 └ secret password
   5   some-app --password $password
   6   echo all done
────────────────────────────────────────


`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			renderer := report.NewSecretRenderer("my-file", test.input, false)
			assert.Equal(t, test.want, strings.ReplaceAll(renderer.Render(), "\r\n", "\n"))
		})
	}
}
