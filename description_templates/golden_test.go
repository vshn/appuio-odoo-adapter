package descriptiontemplates_test

import (
	"context"
	"flag"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/appuio/appuio-cloud-reporting/pkg/invoice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vshn/appuio-odoo-adapter/invoice/desctmpl"
)

const extension = ".gotmpl"

var (
	updateGolden = flag.Bool("update", false, "update the golden files of this test")
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestGenerateGolden(t *testing.T) {
	if *updateGolden {
		require.NoError(t, os.RemoveAll("golden"))
	}

	os.Mkdir("golden", os.ModePerm)

	templateFS := os.DirFS(".")

	fileNames, err := fs.Glob(templateFS, "*"+extension)
	require.NoError(t, err)

	sourceKeys := make([]string, 0, len(fileNames))
	r, err := desctmpl.ItemDescriptionTemplateRendererFromFS(templateFS, extension)
	require.NoError(t, err)
	for _, name := range fileNames {
		if strings.HasPrefix(name, "_") {
			continue
		}
		sourceKeys = append(sourceKeys, strings.TrimSuffix(name, extension))
	}

	sourceKeys = append(sourceKeys, "kafka:exoscale:*:*:business-16", "kafka:exoscale:*:*:premium-30x-32")
	sourceKeys = append(sourceKeys, "opensearch:exoscale:*:*:startup-8", "opensearch:exoscale:*:*:premium-30x-16")
	sourceKeys = append(sourceKeys, "redis:exoscale:*:*:hobbyist-2", "redis:exoscale:*:*:premium-16")
	sourceKeys = append(sourceKeys, "postgres:exoscale:*:*:startup-8", "postgres:exoscale:*:*:premium-32")
	sourceKeys = append(sourceKeys, "postgres:vshn:*:*:standalone-besteffort", "postgres:vshn:*:*:standalone-guaranteed")
	sourceKeys = append(sourceKeys, "mysql:exoscale:*:*:startup-16", "mysql:exoscale:*:*:business-225")

	baseItem := invoice.Item{
		Description: "Long form query description",
		QueryName:   "default_query",
		ProductRef: invoice.ProductRef{
			Target: "1919",
			Source: "SET ME",
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
			"appuio_cloud_memory_subquery_cpu_request": {
				Description: "CPU requests in memory request equivalent",
				QueryName:   "appuio_cloud_memory_subquery_cpu_request",
				Quantity:    44323235.04444221,
				QuantityMin: 2.251,
				QuantityAvg: 133.7,
				QuantityMax: 9001.1,
				Unit:        "TPS",
			},
		},
	}

	for _, key := range sourceKeys {
		t.Run(key, func(t *testing.T) {
			item := baseItem
			item.ProductRef.Source = key

			actual, err := r.RenderItemDescription(context.Background(), item)
			require.NoError(t, err)

			fileName := filepath.Join("golden", key+".txt")
			if *updateGolden {
				require.NoError(t, os.WriteFile(fileName, []byte(actual), os.ModePerm))
				return
			}
			f, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
			require.NoErrorf(t, err, "failed to open golden file %s", fileName)
			defer f.Close()
			expected, err := io.ReadAll(f)
			require.NoErrorf(t, err, "failed to read golden file %s", fileName)

			assert.Equal(t, string(expected), actual)
		})
	}
}
