package invoice

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/appuio/appuio-cloud-reporting/pkg/invoice"
	"github.com/vshn/appuio-odoo-adapter/odoo"
	"github.com/vshn/appuio-odoo-adapter/odoo/model"
)

// CreateInvoice creates a new invoice in Odoo.
func CreateInvoice(ctx context.Context, client *model.Odoo, invoice invoice.Invoice, options ...Option) (int, error) {
	opts := buildOptions(options)

	name := fmt.Sprintf("APPUiO Cloud %s %d", invoice.PeriodStart.Month(), invoice.PeriodStart.Year())
	toCreate := opts.invoiceDefaults
	toCreate.Name = name
	toCreate.Date = odoo.Date(opts.InvoiceDateOrNow())

	partnerID, err := strconv.Atoi(invoice.Tenant.Target)
	if err != nil {
		return 0, fmt.Errorf("error converting tenant target to int: %w", err)
	}
	toCreate.PartnerID = partnerID

	lines := make([]model.InvoiceLine, 0)
	for _, category := range invoice.Categories {
		categoryID, err := strconv.Atoi(category.Target)
		if err != nil {
			return 0, fmt.Errorf("error converting category target to int: %w", err)
		}
		for _, item := range category.Items {
			line := opts.invoiceLineDefaults
			line.CategoryID = categoryID

			total := math.Round(item.Total*100) / 100
			line.Name = fmt.Sprintf("%s Qty: %.2f %s, PPU: %.10f, Disc: %.0f%%",
				item.Description, item.Quantity, item.Unit, item.PricePerUnit, item.Discount*float64(100))

			line.PricePerUnit = total
			line.Quantity = 1
			line.Discount = 0

			productID, err := strconv.Atoi(item.ProductRef.Target)
			if err != nil {
				return 0, fmt.Errorf("error converting product target to int: %w", err)
			}

			line.ProductID = productID

			lines = append(lines, line)
		}
	}

	return createInvoice(ctx, client, toCreate, lines)
}

func createInvoice(ctx context.Context, client *model.Odoo, invoice model.Invoice, lines []model.InvoiceLine) (invoiceID int, err error) {
	created, err := client.CreateInvoice(ctx, invoice)
	if err != nil {
		return created.ID, fmt.Errorf("error creating invoice in odoo: %w", err)
	}

	createdLines := make([]model.InvoiceLine, 0, len(lines))
	for _, line := range lines {
		line, err := client.InvoiceAddLine(ctx, created.ID, line)
		createdLines = append(createdLines, line)
		if err != nil {
			return created.ID, fmt.Errorf("error adding line to invoice %d: %w; created until error %+v", created.ID, err, createdLines)
		}
	}

	err = client.InvoiceCalculateTaxes(ctx, created.ID)
	if err != nil {
		return created.ID, fmt.Errorf("error calculating taxes on invoice %d: %w", created.ID, err)
	}

	return created.ID, nil
}
