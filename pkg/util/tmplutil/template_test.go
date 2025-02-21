package tmplutil

import (
	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/assert"
	strings2 "github.com/weiliang-ms/easyctl/pkg/util/strings"
	"testing"
	"text/template"
)

func TestRender(t *testing.T) {
	tmpl := template.Must(template.New("tmpl-demo").Parse(dedent.Dedent(`
{{- if .Ports }}
{{ range .Ports }}
echo {{ . }}
{{- if $.Password }}
echo {{ $.Password }}
{{- end }}
{{- end }}
{{- end }}
`)))
	content, err := Render(tmpl, TmplRenderData{
		"Ports":    []int{2, 3, 4},
		"Password": "password",
	})

	const expect = `
echo 2
echo password
echo 3
echo password
echo 4
echo password`

	assert.Nil(t, err)
	assert.Equal(t, expect, strings2.TrimPrefixAndSuffix(content, "\n"))

	_ , err = Render(&template.Template{}, TmplRenderData{})
	assert.NotNil(t, err)
}
