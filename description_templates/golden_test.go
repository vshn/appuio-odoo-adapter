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

	baseItem := invoice.Item{
		Description: "Long form query description",
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
		SubItems: []invoice.SubItem{
			{
				Description: "Memory requests",
				Quantity:    34923234.04433424,
				QuantityMin: 2.251,
				QuantityAvg: 42.2,
				QuantityMax: 9001.1,
				Unit:        "TPS",
			},
			{
				Description: "CPU requests in memory request equivalent",
				Quantity:    34923234.04433424,
				QuantityMin: 2.251,
				QuantityAvg: 42.2,
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
