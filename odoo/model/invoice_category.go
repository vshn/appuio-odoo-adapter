package model

import (
	"context"

	"github.com/vshn/appuio-odoo-adapter/odoo"
)

// InvoiceCategory (alias "Section" in Invoices) visually categorizes line items into logical groups.
type InvoiceCategory struct {
	// ID is the data record identifier.
	ID        int    `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Sequence  int    `json:"sequence,omitempty"`
	PageBreak bool   `json:"pagebreak,omitempty"`
	Separator bool   `json:"separator,omitempty"`
	SubTotal  bool   `json:"subtotal,omitempty"`
}

// CreateCategory creates a new invoice category and returns the ID of the data record.
// Note that setting InvoiceCategory.ID in the payload doesn't have an effect.
func (o Odoo) CreateCategory(ctx context.Context, category InvoiceCategory) (int, error) {
	return o.client.CreateGenericModel(ctx, o.session, odoo.NewCreateModel("sale_layout.category", category))
}

// UpdateCategory updates a given invoice category and returns true if the data record has been updated.
func (o Odoo) UpdateCategory(ctx context.Context, category InvoiceCategory) (bool, error) {
	m, err := odoo.NewUpdateModel("sale_layout.category", category.ID, category)
	if err != nil {
		return false, err
	}
	return o.client.UpdateGenericModel(ctx, o.session, m)
}

// FetchCategoryByID searches for the invoice category by ID and returns the first entry in the result.
// If no result has been found, nil is returned without error.
func (o Odoo) FetchCategoryByID(ctx context.Context, id int) (*InvoiceCategory, error) {
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

// SearchCategoriesByName searches for invoice categories with the given name.
// If no results have been found, an empty slice is returned without error.
func (o Odoo) SearchCategoriesByName(ctx context.Context, name string) ([]InvoiceCategory, error) {
	// TODO: Set filters
	return o.searchCategories(ctx, []odoo.Filter{})
}

func (o Odoo) searchCategories(ctx context.Context, domainFilters []odoo.Filter) ([]InvoiceCategory, error) {
	type readResult struct {
		Records []InvoiceCategory `json:"records"`
	}
	result := &readResult{}

	err := o.client.SearchGenericModel(ctx, o.session, odoo.SearchReadModel{
		Model:  "sale_layout.category",
		Domain: domainFilters,
		Fields: []string{"name", "sequence", "pagebreak", "separator", "subtotal"},
	}, result)
	return result.Records, err
}
