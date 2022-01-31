package invoice

import (
	"time"

	"github.com/vshn/appuio-odoo-adapter/odoo/model"
)

type options struct {
	invoiceDate time.Time

	invoiceDefaults     model.Invoice
	invoiceLineDefaults model.InvoiceLine
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
