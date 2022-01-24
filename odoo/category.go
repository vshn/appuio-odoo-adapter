package odoo

import (
	"context"
)

// Category (alias "Section" in Invoices) visually categorizes line items into logical groups.
type Category struct {
	// ID is the data record identifier.
	ID        int    `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Sequence  int    `json:"sequence,omitempty"`
	PageBreak bool   `json:"pagebreak,omitempty"`
	Separator bool   `json:"separator,omitempty"`
	SubTotal  bool   `json:"subtotal,omitempty"`
}

// CreateCategory creates a new invoice category and returns the ID of the data record.
// Note that setting Category.ID in the payload doesn't have an effect.
func (c Client) CreateCategory(ctx context.Context, session *Session, category Category) (int, error) {
	return c.CreateGenericModel(ctx, session, NewCreateModel("sale_layout.category", category))
}

// UpdateCategory updates a given invoice category and returns true if the data record has been updated.
func (c Client) UpdateCategory(ctx context.Context, session *Session, category Category) (bool, error) {
	m, err := NewUpdateModel("sale_layout.category", category.ID, category)
	if err != nil {
		return false, err
	}
	return c.UpdateGenericModel(ctx, session, m)
}

// FetchCategoryByID searches for the invoice category by ID and returns the first entry in the result.
// If no result has been found, nil is returned without error.
func (c Client) FetchCategoryByID(ctx context.Context, session *Session, id int) (*Category, error) {
	result, err := c.searchCategories(ctx, session, []Filter{
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
func (c Client) SearchCategoriesByName(ctx context.Context, session *Session, name string) ([]Category, error) {
	// TODO: Set filters
	return c.searchCategories(ctx, session, []Filter{})
}

func (c Client) searchCategories(ctx context.Context, session *Session, domainFilters []Filter) ([]Category, error) {
	type readResult struct {
		Records []Category `json:"records"`
	}
	result := &readResult{}

	err := c.SearchGenericModel(ctx, session, SearchReadModel{
		Model:  "sale_layout.category",
		Domain: domainFilters,
		Fields: []string{"name", "sequence", "pagebreak", "separator", "subtotal"},
	}, result)
	return result.Records, err
}
