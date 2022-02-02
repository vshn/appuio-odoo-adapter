package invoice

import (
	"context"
	"fmt"
	"time"

	"github.com/appuio/appuio-cloud-reporting/pkg/invoice"
	"github.com/vshn/appuio-odoo-adapter/odoo/model"
)

type options struct {
	invoiceDate time.Time

	invoiceDefaults     model.Invoice
	invoiceLineDefaults model.InvoiceLine

	itemDescriptionRenderer ItemDescriptionRenderer
}

// Option represents a report option.
type Option interface {
	set(*options)
}

func buildOptions(os []Option) options {
	var build options
	for _, o := range os {
		o.set(&build)
	}
	return build
}

func (o options) InvoiceDateOrNow() time.Time {
	if o.invoiceDate.IsZero() {
		return time.Now()
	}
	return o.invoiceDate
}

func (o options) ItemDescriptionRenderer() ItemDescriptionRenderer {
	if o.itemDescriptionRenderer != nil {
		return o.itemDescriptionRenderer
	}

	return DefaultItemDescriptionRenderer{}
}

// WithInvoiceDate sets the date for the invoice.
func WithInvoiceDate(tm time.Time) Option {
	return invoiceDate(tm)
}

type invoiceDate time.Time

func (t invoiceDate) set(o *options) {
	o.invoiceDate = time.Time(t)
}

// WithInvoiceDefaults sets the defaults for a invoice.
func WithInvoiceDefaults(tm model.Invoice) Option {
	return invoiceDefaults(tm)
}

type invoiceDefaults model.Invoice

func (t invoiceDefaults) set(o *options) {
	o.invoiceDefaults = model.Invoice(t)
}

// WithInvoiceLineDefaults sets the defaults for an invoice line.
func WithInvoiceLineDefaults(tm model.InvoiceLine) Option {
	return invoiceLineDefaults(tm)
}

type invoiceLineDefaults model.InvoiceLine

func (t invoiceLineDefaults) set(o *options) {
	o.invoiceLineDefaults = model.InvoiceLine(t)
}

type ItemDescriptionRenderer interface {
	RenderItemDescription(context.Context, invoice.Item) (string, error)
}

// WithInvoiceLineDescriptionTemplate sets the description for an invoice line.
func WithItemDescriptionRenderer(tm ItemDescriptionRenderer) Option {
	return itemDescriptionRendererOpt{tm}
}

type itemDescriptionRendererOpt struct{ ItemDescriptionRenderer }

func (t itemDescriptionRendererOpt) set(o *options) {
	o.itemDescriptionRenderer = t.ItemDescriptionRenderer
}

// DefaultItemDescriptionRenderer is the default way to render an item description.
type DefaultItemDescriptionRenderer struct{}

// RenderItemDescription renders the item description using the extended default format "%+v"
func (r DefaultItemDescriptionRenderer) RenderItemDescription(_ context.Context, i invoice.Item) (string, error) {
	return fmt.Sprintf("%+v", i), nil
}
