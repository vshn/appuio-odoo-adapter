package desctmpl

import (
	"context"
	"fmt"
	"io/fs"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"

	"github.com/appuio/appuio-cloud-reporting/pkg/invoice"
)

var rootTemplate = template.New("root").Funcs(template.FuncMap{
	"perMinute": func(v float64) float64 { return v / float64(60) },
}).Funcs(sprig.TxtFuncMap())

// ItemDescriptionTemplateRenderer renders item descriptions using the golang template engine.
// It uses the `.ProductRef.Source` as key to look the template up.
// Sprig helpers are included.
type ItemDescriptionTemplateRenderer struct {
	extension string

	root *template.Template
}

// ItemDescriptionTemplateRendererFromFS creates a new ItemDescriptionTemplateRenderer from the file system.
func ItemDescriptionTemplateRendererFromFS(fs fs.FS, extension string) (*ItemDescriptionTemplateRenderer, error) {
	root, err := rootTemplate.Clone()
	if err != nil {
		return nil, err
	}
	_, err = root.ParseFS(fs, "*"+extension)
	if err != nil {
		return nil, err
	}

	return &ItemDescriptionTemplateRenderer{extension, root}, nil
}

// RenderItemDescription renders an item description. Uses the `.ProductRef.Source` as the key to look which template to use.
func (r *ItemDescriptionTemplateRenderer) RenderItemDescription(_ context.Context, item invoice.Item) (string, error) {
	key := item.ProductRef.Source
	tmpl := r.root.Lookup(key + r.extension)
	if tmpl == nil {
		return "", fmt.Errorf("failed to find template for `ProductRef.Source=%q`%s", key, r.root.DefinedTemplates())
	}

	b := &strings.Builder{}
	err := r.root.ExecuteTemplate(b, item.ProductRef.Source+r.extension, item)
	return b.String(), err
}
