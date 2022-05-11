package descriptiontemplates_test

import (
	"context"
	"os"
	"testing"

	"github.com/appuio/appuio-cloud-reporting/pkg/invoice"
	"github.com/stretchr/testify/require"
	"github.com/vshn/appuio-odoo-adapter/invoice/desctmpl"
)

func TestTempalte_MissingSubItem_NoError(t *testing.T) {
	templateFS := os.DirFS(".")
	r, err := desctmpl.ItemDescriptionTemplateRendererFromFS(templateFS, extension)
	require.NoError(t, err)

	item := invoice.Item{
		Description: "Long form query description",
		QueryName:   "default_query",
		ProductRef: invoice.ProductRef{
			Target: "1919",
			Source: "appuio_cloud_memory:c-appuio-exoscale-ch-gva-2-0",
		},
		Quantity:     87955674.09456764,
		QuantityMin:  456.345593,
		QuantityAvg:  456.345593,
		QuantityMax:  456.345593,
		Unit:         "UNIT",
		PricePerUnit: 0.000000746,
		Discount:     0.33,
		Total:        43.962005025946798,
		SubItems: map[string]invoice.SubItem{
			"appuio_cloud_memory_subquery_memory_request": {
				Description: "Memory request aggregated by namespace",
				QueryName:   "appuio_cloud_memory_subquery_memory_request",
				Quantity:    34923234.04433424,
				QuantityMin: 2.251,
				QuantityAvg: 42.2,
				QuantityMax: 9001.1,
				Unit:        "TPS",
			},
		},
	}

	_, err = r.RenderItemDescription(context.Background(), item)
	require.NoError(t, err)
}
