package invoice_test

import (
	"context"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/appuio/appuio-cloud-reporting/pkg/invoice"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	. "github.com/vshn/appuio-odoo-adapter/invoice"
	"github.com/vshn/appuio-odoo-adapter/odoo"
	"github.com/vshn/appuio-odoo-adapter/odoo/model"
	"github.com/vshn/appuio-odoo-adapter/odoo/odoomock"
)

func TestOdooInvoiceCreator_CreateInvoice(t *testing.T) {
	invoiceDate := time.Now()

	invoiceDefaults := model.Invoice{
		AccountID: 666,
	}
	invoiceLineDefaults := model.InvoiceLine{
		AccountID: 7666,
	}

	subject := invoice.Invoice{
		PeriodStart: time.Date(2021, time.December, 1, 0, 0, 0, 0, time.UTC),
		Tenant: invoice.Tenant{
			Source: "umbrellacorp",
			Target: "19680000",
		},
		Categories: []invoice.Category{
			{Source: "us-rac-2:disposal-plant-p-12a-furnace-control", Target: "19680010", Items: []invoice.Item{
				{
					Description:  "APPUiO Cloud Memory",
					ProductRef:   invoice.ProductRef{Target: "660"},
					Unit:         "MiB",
					Quantity:     1483434.78,
					PricePerUnit: 0.0000078,
					Discount:     0,
					Total:        148.78 * 0.0000078,
				},
			}},
			{Source: "us-rac-2:nest-elevator-control", Target: "19680020", Items: []invoice.Item{
				{
					Description: "APPUiO Cloud Memory",
					ProductRef:  invoice.ProductRef{Target: "660"},

					Unit:         "MiB",
					Quantity:     2455,
					PricePerUnit: 0.0000078,
					Discount:     0,
					Total:        2455 * 0.0000078,
				}, {
					Description:  "APPUiO Cloud RWX Storage",
					ProductRef:   invoice.ProductRef{Target: "810"},
					Unit:         "GiB",
					Quantity:     100,
					PricePerUnit: 0.00034,
					Discount:     0.2,
					Total:        148.78 * 0.0000078 * 0.8,
				},
			}},
		},
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockExecutor := odoomock.NewMockQueryExecutor(mockCtrl)

	round := func(f float64) float64 { return math.Round(f*100) / 100 }
	renderDesc := func(i invoice.Item) string {
		s, _ := DefaultItemDescriptionRenderer{}.RenderItemDescription(context.Background(), i)
		return s
	}

	// TODO directly mock api client instead of the underlying executor
	invCreateCall := mockExecutor.
		EXPECT().
		CreateGenericModel(context.Background(), "account.invoice", func() model.Invoice {
			inv := invoiceDefaults
			inv.Date = odoo.Date(invoiceDate)

			inv.PartnerID, _ = strconv.Atoi(subject.Tenant.Target)
			inv.Name = "APPUiO Cloud December 2021"
			return inv
		}())

	line1Create := mockExecutor.
		EXPECT().
		CreateGenericModel(context.Background(), "account.invoice.line", func() model.InvoiceLine {
			line := invoiceLineDefaults
			line.CategoryID, _ = strconv.Atoi(subject.Categories[0].Target)

			line.Name = renderDesc(subject.Categories[0].Items[0])
			line.PricePerUnit = round(subject.Categories[0].Items[0].Total)
			line.Quantity = 1
			line.Discount = 0

			line.ProductID, _ = strconv.Atoi(subject.Categories[0].Items[0].ProductRef.Target)
			return line
		}())
	line2Create := mockExecutor.
		EXPECT().
		CreateGenericModel(context.Background(), "account.invoice.line", func() model.InvoiceLine {
			line := invoiceLineDefaults
			line.CategoryID, _ = strconv.Atoi(subject.Categories[1].Target)

			line.Name = renderDesc(subject.Categories[1].Items[0])
			line.PricePerUnit = round(subject.Categories[1].Items[0].Total)
			line.Quantity = 1
			line.Discount = 0

			line.ProductID, _ = strconv.Atoi(subject.Categories[1].Items[0].ProductRef.Target)
			return line
		}())
	line3Create := mockExecutor.
		EXPECT().
		CreateGenericModel(context.Background(), "account.invoice.line", func() model.InvoiceLine {
			line := invoiceLineDefaults
			line.CategoryID, _ = strconv.Atoi(subject.Categories[1].Target)

			line.Name = renderDesc(subject.Categories[1].Items[1])
			line.PricePerUnit = round(subject.Categories[1].Items[1].Total)
			line.Quantity = 1
			line.Discount = 0

			line.ProductID, _ = strconv.Atoi(subject.Categories[1].Items[1].ProductRef.Target)
			return line
		}())

	calculateTaxCall := mockExecutor.
		EXPECT().
		ExecuteQuery(context.Background(), "/web/dataset/call_kw/button_reset_taxes", gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, _ interface{}, ok *bool) error {
			*ok = true
			return nil
		})

	gomock.InOrder(
		invCreateCall,
		line1Create,
		line2Create,
		line3Create,
		calculateTaxCall,
	)

	_, err := CreateInvoice(context.Background(), model.NewOdoo(mockExecutor), subject,
		WithInvoiceDate(invoiceDate),
		WithInvoiceDefaults(invoiceDefaults),
		WithInvoiceLineDefaults(invoiceLineDefaults),
	)
	require.NoError(t, err)
}
