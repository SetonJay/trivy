package report

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/liamg/tml"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/aquasecurity/trivy/pkg/fanal/types"
)

type secretRenderer struct {
	target  string
	secrets []types.SecretFinding
	w       *bytes.Buffer
	width   int
	ansi    bool
}

func NewSecretRenderer(target string, secrets []types.SecretFinding, ansi bool) *secretRenderer {
	width, _, err := terminal.GetSize(0)
	if err != nil || width == 0 {
		width = 40
	}
	if !ansi {
		tml.DisableFormatting()
	}
	return &secretRenderer{
		w:       bytes.NewBuffer([]byte{}),
		target:  target,
		secrets: secrets,
		width:   width,
		ansi:    ansi,
	}
}

func (r *secretRenderer) Render() string {
	for _, m := range r.secrets {
		r.renderSingle(m)
	}
	return r.w.String()
}

func (r *secretRenderer) printf(format string, args ...interface{}) {
	// nolint
	_ = tml.Fprintf(r.w, format, args...)
}

func (r *secretRenderer) printDoubleDivider() {
	r.printf("<dim>%s\r\n", strings.Repeat("═", r.width))
}

func (r *secretRenderer) printSingleDivider() {
	r.printf("<dim>%s\r\n", strings.Repeat("─", r.width))
}

func (r *secretRenderer) renderSingle(secret types.SecretFinding) {
	r.renderSummary(secret)
	r.renderCode(secret)
	r.printf("\r\n\r\n")
}

func (r *secretRenderer) renderSummary(secret types.SecretFinding) {

	// severity
	switch secret.Severity {
	case severityCritical:
		r.printf("<red><bold>%s: ", secret.Severity)
	case severityHigh:
		r.printf("<red>%s: ", secret.Severity)
	case severityMedium:
		r.printf("<yellow>%s: ", secret.Severity)
	case severityLow:
		r.printf("%s: ", secret.Severity)
	default:
		r.printf("<blue>%s: ", secret.Severity)
	}

	// heading
	r.printf("%s (%s)\r\n", secret.Category, secret.RuleID)
	r.printDoubleDivider()

	// description
	r.printf("<dim>%s\r\n", secret.Title)

	r.printSingleDivider()
}

func (r *secretRenderer) renderCode(secret types.SecretFinding) {
	// highlight code if we can...
	if lines := secret.Code.Lines; len(lines) > 0 {

		var lineInfo string
		if secret.StartLine > 0 {
			lineInfo = tml.Sprintf("<dim>:</dim><blue>%d", secret.StartLine)
			if secret.EndLine > secret.StartLine {
				lineInfo = tml.Sprintf("%s<blue>-%d", lineInfo, secret.EndLine)
			}
		}
		r.printf(" <blue>%s%s\r\n", r.target, lineInfo)
		r.printSingleDivider()
		for i, line := range lines {
			if line.Truncated {
				r.printf("<dim>%4s   ", strings.Repeat(".", len(fmt.Sprintf("%d", line.Number))))
			} else if line.IsCause {
				r.printf("<red>%4d ", line.Number)
				switch {
				case (line.FirstCause && line.LastCause) || len(lines) == 1:
					r.printf("<red>[ ")
				case line.FirstCause || i == 0:
					r.printf("<red>┌ ")
				case line.LastCause || i == len(lines)-1:
					r.printf("<red>└ ")
				default:
					r.printf("<red>│ ")
				}
			} else {
				r.printf("<dim>%4d   ", line.Number)
			}
			if r.ansi {
				r.printf("%s\r\n", line.Highlighted)
			} else {
				r.printf("%s\r\n", line.Content)
			}
		}
		r.printSingleDivider()
	}
}
