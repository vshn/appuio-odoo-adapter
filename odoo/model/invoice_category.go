package model

import (
	"context"

	"github.com/vshn/appuio-odoo-adapter/odoo"
)

// InvoiceCategory (alias "Section" in Invoices) visually categorizes line items into logical groups.
type InvoiceCategory struct {
	// ID is the data record identifier.
	ID int `json:"id,omitempty"`
	// Name is the title of the category/section within an invoice.
	Name string `json:"name,omitempty"`
	// Sequence controls how line items are grouped together and how each category is ordered within an invoice.
	//
	// If there are line items that share the same category, but each category has a unique sequence number, then all line items are grouped together.
	// In this case the order of line items in the invoice doesn't matter.
	// The ordering among the categories themselves is ascending with the category with the lowest sequence number appearing at the top.
	//
	// If there are categories that share the same sequence number then the ordering of line items greatly matters within an invoice.
	// Line items that share the same category but aren't sequentially defined in the invoice end up distributed ("as is") with the category appearing multiple times.
	Sequence int `json:"sequence,omitempty"`
	// PageBreak causes the next line item to be on the next page in a PDF (after the last line item).
	PageBreak bool `json:"pagebreak,omitempty"`
	// Separator causes a line printed with "***" after the last line item within the same category.
	Separator bool `json:"separator,omitempty"`
	// SubTotal causes an additional line printed with an accumulated subtotal of all line items within the same category.
	SubTotal bool `json:"subtotal,omitempty"`
}

// CreateInvoiceCategory creates a new invoice category and returns the ID of the data record.
// Note that setting InvoiceCategory.ID in the payload doesn't have an effect.
func (o Odoo) CreateInvoiceCategory(ctx context.Context, category InvoiceCategory) (int, error) {
	return o.session.CreateGenericModel(ctx, "sale_layout.category", category)
}

// UpdateInvoiceCategory updates a given invoice category and returns true if the data record has been updated.
func (o Odoo) UpdateInvoiceCategory(ctx context.Context, category InvoiceCategory) (bool, error) {
	return o.session.UpdateGenericModel(ctx, "sale_layout.category", category.ID, category)
}

// DeleteInvoiceCategory updates a given invoice category and returns true if the data record has been updated.
// For all existing invoices, the "section" field of all affected line items become empty.
func (o Odoo) DeleteInvoiceCategory(ctx context.Context, category InvoiceCategory) (bool, error) {
	return o.session.DeleteGenericModel(ctx, "sale_layout.category", []int{category.ID})
}

// FetchInvoiceCategoryByID searches for the invoice category by ID and returns the first entry in the result.
// If no result has been found, nil is returned without error.
func (o Odoo) FetchInvoiceCategoryByID(ctx context.Context, id int) (*InvoiceCategory, error) {
	result, err := o.searchCategories(ctx, []odoo.Filter{
		[]interface{}{"id", "in", []int{id}},
	})
	if err != nil {
		return nil, err
	}
	if len(result) > 0 {
		return &result[0], nil
	}
	// not found
	return nil, nil
}

// SearchInvoiceCategoriesByName searches for invoice categories that include the given string.
// The search is case-insensitive.
// If no results have been found, an empty slice is returned without error.
func (o Odoo) SearchInvoiceCategoriesByName(ctx context.Context, searchString string) ([]InvoiceCategory, error) {
	return o.searchCategories(ctx, []odoo.Filter{
		[]string{"name", "ilike", searchString},
	})
}

func (o Odoo) searchCategories(ctx context.Context, domainFilters []odoo.Filter) ([]InvoiceCategory, error) {
	type readResult struct {
		Records []InvoiceCategory `json:"records"`
	}
	result := &readResult{}

	err := o.session.SearchGenericModel(ctx, odoo.SearchReadModel{
		Model:  "sale_layout.category",
		Domain: domainFilters,
		Fields: []string{"name", "sequence", "pagebreak", "separator", "subtotal"},
	}, result)
	return result.Records, err
}
