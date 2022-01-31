package model

import (
	"context"
	"fmt"

	"github.com/vshn/appuio-odoo-adapter/odoo"
)

// Invoice represents an Odoo invoice.
type Invoice struct {
	// ID is the data record identifier.
	ID int `json:"id,omitempty" yaml:"id,omitempty"`

	// Name is the title of the invoice shown in odoo and the field "Beschreibung/Reference" in the PDF
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Date is the date the invoice date.
	Date odoo.Date `json:"date_invoice,omitempty" yaml:"date_invoice,omitempty"`
	// State reprents the state of the invoice. Known states include: [draft, proforma2, open, cancel, paid]
	State string `json:"state,omitempty" yaml:"state,omitempty"`
	// UserID is the id of the user that initially created the invoice.
	UserID int `json:"user_id,omitempty" yaml:"user_id,omitempty"`

	// PaymentTermID is the id of the payment terms used e.g. 10 days or 30 days
	PaymentTermID int `json:"payment_term,omitempty" yaml:"payment_term,omitempty"`
	// AccountID is the id of the account. The account is something like "1100 Forderungen ggü. Dritten aus der Schweiz".
	AccountID int `json:"account_id,omitempty" yaml:"account_id,omitempty"`
	// CurrencyID is the id of the used currency.
	CurrencyID int `json:"currency_id,omitempty" yaml:"currency_id,omitempty"`
	// JournalID is the journal (or book) id.
	JournalID int `json:"journal_id,omitempty" yaml:"journal_id,omitempty"`
	// PartnerID is the partner (or customer) id.
	PartnerID int `json:"partner_id,omitempty" yaml:"partner_id,omitempty"`
}

// InvoiceLine represents a line in the Odoo invoice.
type InvoiceLine struct {
	// ID is the data record identifier.
	ID int `json:"id,omitempty" yaml:"id,omitempty"`

	// InvoiceID is the id of the referenced invoice.
	InvoiceID int `json:"invoice_id,omitempty" yaml:"invoice_id,omitempty"`

	// Name is the description of the invoice shown in odoo and the field "Beschreibung" in the PDF.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Sequence controls how line items are grouped together and how each category is ordered within an invoice.
	//
	// If there are line items that share the same category, but each category has a unique sequence number, then all line items are grouped together.
	// In this case the order of line items in the invoice doesn't matter.
	// The ordering among the categories themselves is ascending with the category with the lowest sequence number appearing at the top.
	//
	// If there are categories that share the same sequence number then the ordering of line items greatly matters within an invoice.
	// Line items that share the same category but aren't sequentially defined in the invoice end up distributed ("as is") with the category appearing multiple times.
	Sequence int `json:"sequence" yaml:"sequence"`

	// PricePerUnit is the price per unit.
	PricePerUnit float64 `json:"price_unit" yaml:"price_unit"`
	// Quantity is the amount of units.
	Quantity float64 `json:"quantity" yaml:"quantity"`
	// Discount is the discount in percent.
	Discount int `json:"discount" yaml:"discount"`

	// AccountID is the id of the account. The account is something like "3400 Dienstleistungserlöse".
	AccountID int `json:"account_id,omitempty" yaml:"account_id,omitempty"`
	// ProductID is the id of the product.
	ProductID int `json:"product_id,omitempty" yaml:"product_id,omitempty"`
	// CategoryID is the id of the category. See InvoiceCategory for more information.
	CategoryID int `json:"sale_layout_cat_id,omitempty" yaml:"sale_layout_cat_id,omitempty"`
	// TaxID represents the id of the VAT.
	TaxID []InvoiceLineTaxID `json:"invoice_line_tax_id,omitempty" yaml:"invoice_line_tax_id,omitempty"`
}

// CreateInvoice creates a new invoice.
func (o *Odoo) CreateInvoice(ctx context.Context, inv Invoice) (Invoice, error) {
	n, err := o.querier.CreateGenericModel(ctx, "account.invoice", inv)
	inv.ID = n
	if err != nil {
		return inv, fmt.Errorf("error while creating an invoice: %w", err)
	}
	return inv, nil
}

// InvoiceCalculateTaxes calculates taxes on an invoice.
func (o *Odoo) InvoiceCalculateTaxes(ctx context.Context, invoiceID int) error {
	var ok bool
	err := o.querier.ExecuteQuery(ctx, "/web/dataset/call_kw/button_reset_taxes", odoo.WriteModel{
		Model:  "account.invoice",
		Method: "button_reset_taxes",
		Args: []interface{}{
			[]int{invoiceID},
		},
		KWArgs: map[string]interface{}{}, // set to non-null when serializing
	}, &ok)

	if err == nil && !ok {
		err = fmt.Errorf("expected odoo to return %t got %t", true, ok)
	}
	if err != nil {
		return fmt.Errorf("error calculating taxes on invoice %d: %w", invoiceID, err)
	}
	return nil
}

// InvoiceAddLine adds a line to the invoice with the given id.
func (o *Odoo) InvoiceAddLine(ctx context.Context, invoiceID int, line InvoiceLine) (InvoiceLine, error) {
	line.InvoiceID = invoiceID
	n, err := o.querier.CreateGenericModel(ctx, "account.invoice.line", line)
	line.ID = n
	if err != nil {
		return line, fmt.Errorf("error while adding line to invoice: %w", err)
	}
	return line, nil
}
