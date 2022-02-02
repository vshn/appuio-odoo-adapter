package desctmpl_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/appuio/appuio-cloud-reporting/pkg/invoice"
	"github.com/stretchr/testify/require"
	"github.com/vshn/appuio-odoo-adapter/invoice/desctmpl"
)

func TestRenderItemDescription(t *testing.T) {
	extension := ".gotmpl"
	templateFS := fstest.MapFS{
		"memory" + extension: &fstest.MapFile{
			Data: []byte("{{.ProductRef.Source}}: {{.Total}}"),
		},
		"storage" + extension: &fstest.MapFile{
			Data: []byte("so vieli bytesli: {{.Total}}"),
		},
	}

	subject, err := desctmpl.ItemDescriptionTemplateRendererFromFS(templateFS, extension)
	require.NoError(t, err)

	tests := []struct {
		desc        string
		item        invoice.Item
		expectedOut string
		expectedErr require.ErrorAssertionFunc
	}{
		{
			"memory source",
			invoice.Item{ProductRef: invoice.ProductRef{Source: "memory"}, Total: 77},
			"memory: 77",
			require.NoError,
		}, {
			"storage source",
			invoice.Item{ProductRef: invoice.ProductRef{Source: "storage"}, Total: 99},
			"so vieli bytesli: 99",
			require.NoError,
		}, {
			"unknown source",
			invoice.Item{ProductRef: invoice.ProductRef{Source: "unknown"}, Total: 77},
			"",
			require.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			rendered, err := subject.RenderItemDescription(context.Background(), tc.item)
			require.Equal(t, tc.expectedOut, rendered)
			tc.expectedErr(t, err)
		})
	}
}
