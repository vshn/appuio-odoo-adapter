package invoice_test

import (
	"context"
	"fmt"
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

	partnerId := 19680000
	subject := invoice.Invoice{
		PeriodStart: time.Date(2021, time.December, 1, 0, 0, 0, 0, time.UTC),
		Tenant: invoice.Tenant{
			Source: "umbrellacorp",
			Target: strconv.FormatInt(int64(partnerId), 10),
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

	gomock.InOrder(
		mockPartnerQueryCall(mockExecutor, model.Partner{ID: partnerId, Name: "Umbrella Corp Ltd."}),
		mockInvoiceCreateCall(mockExecutor, invoiceDefaults, invoiceDate, partnerId, "Umbrella Corp Ltd. APPUiO Cloud December 2021"),
		mockInvoiceLineCreateCall(mockExecutor, invoiceLineDefaults, subject.Categories[0], subject.Categories[0].Items[0]),
		mockInvoiceLineCreateCall(mockExecutor, invoiceLineDefaults, subject.Categories[1], subject.Categories[1].Items[0]),
		mockInvoiceLineCreateCall(mockExecutor, invoiceLineDefaults, subject.Categories[1], subject.Categories[1].Items[1]),
		mockCalculateTaxCall(mockExecutor),
	)

	_, err := CreateInvoice(context.Background(), model.NewOdoo(mockExecutor), subject,
		WithInvoiceDate(invoiceDate),
		WithInvoiceDefaults(invoiceDefaults),
		WithInvoiceLineDefaults(invoiceLineDefaults),
	)
	require.NoError(t, err)
}

func TestOdooInvoiceCreator_CreateInvoiceWithParentID(t *testing.T) {
	invoiceDate := time.Now()

	invoiceDefaults := model.Invoice{
		AccountID: 666,
	}
	invoiceLineDefaults := model.InvoiceLine{
		AccountID: 7666,
	}

	partnerId := 19680000
	subject := invoice.Invoice{
		PeriodStart: time.Date(2021, time.December, 1, 0, 0, 0, 0, time.UTC),
		Tenant: invoice.Tenant{
			Source: "umbrellacorp",
			Target: strconv.FormatInt(int64(partnerId), 10),
		},
		Categories: []invoice.Category{},
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockExecutor := odoomock.NewMockQueryExecutor(mockCtrl)

	gomock.InOrder(
		mockPartnerQueryCall(mockExecutor, model.Partner{
			ID:     1111111111111111111,
			Name:   "Umbrella Corp Ltd. Billing Department",
			Parent: model.OdooCompositeID{Valid: true, ID: 19680000, Name: "Umbrella Corp Ltd."},
		}),
		mockInvoiceCreateCall(mockExecutor, invoiceDefaults, invoiceDate, partnerId, "Umbrella Corp Ltd. APPUiO Cloud December 2021"),
		mockCalculateTaxCall(mockExecutor),
	)

	_, err := CreateInvoice(context.Background(), model.NewOdoo(mockExecutor), subject,
		WithInvoiceDate(invoiceDate),
		WithInvoiceDefaults(invoiceDefaults),
		WithInvoiceLineDefaults(invoiceLineDefaults),
	)
	require.NoError(t, err)
}

func mockPartnerQueryCall(mockExecutor *odoomock.MockQueryExecutor, partner model.Partner) *gomock.Call {
	return mockExecutor.
		EXPECT().
		SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).
		Do(func(ctx context.Context, s odoo.SearchReadModel, into interface{}) error {
			pl, ok := into.(*model.PartnerList)
			if !ok {
				return fmt.Errorf("Expected into to be of type *model.PartnerList")
			}
			pl.Items = append(pl.Items, partner)
			return nil
		})
}

func mockInvoiceCreateCall(mockExecutor *odoomock.MockQueryExecutor, defaults model.Invoice, date time.Time, partnerId int, name string) *gomock.Call {
	return mockExecutor.
		EXPECT().
		CreateGenericModel(context.Background(), "account.invoice", func() model.Invoice {
			inv := defaults
			inv.Date = odoo.Date(date)

			inv.PartnerID = partnerId
			inv.Name = name
			return inv
		}())
}

func mockCalculateTaxCall(mockExecutor *odoomock.MockQueryExecutor) *gomock.Call {
	return mockExecutor.
		EXPECT().
		ExecuteQuery(context.Background(), "/web/dataset/call_kw/button_reset_taxes", gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, _ interface{}, ok *bool) error {
			*ok = true
			return nil
		})
}

func mockInvoiceLineCreateCall(mockExecutor *odoomock.MockQueryExecutor, defaults model.InvoiceLine, category invoice.Category, item invoice.Item) *gomock.Call {
	round := func(f float64) float64 { return math.Round(f*100) / 100 }
	renderDesc := func(i invoice.Item) string {
		s, _ := DefaultItemDescriptionRenderer{}.RenderItemDescription(context.Background(), i)
		return s
	}

	return mockExecutor.
		EXPECT().
		CreateGenericModel(context.Background(), "account.invoice.line", func() model.InvoiceLine {
			line := defaults
			line.CategoryID, _ = strconv.Atoi(category.Target)

			line.Name = renderDesc(item)
			line.PricePerUnit = round(item.Total)
			line.Quantity = 1
			line.Discount = 0

			line.ProductID, _ = strconv.Atoi(item.ProductRef.Target)
			return line
		}())
}
