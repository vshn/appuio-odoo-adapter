package odoo

import (
	"context"
)

// Category (alias "Section" in Invoices) visually categorizes line items into logical groups.
type Category struct {
	ID        int    `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Sequence  int    `json:"sequence,omitempty"`
	PageBreak bool   `json:"pagebreak,omitempty"`
	Separator bool   `json:"separator,omitempty"`
	SubTotal  bool   `json:"subtotal,omitempty"`
}

func (c Client) CreateCategory(ctx context.Context, session *Session, category Category) (int, error) {
	return c.CreateGenericModel(ctx, session, NewCreateModel("sale_layout.category", category))
}

func (c Client) UpdateCategory(ctx context.Context, session *Session, category Category) (bool, error) {
	m, err := NewUpdateModel("sale_layout.category", category.ID, category)
	if err != nil {
		return false, err
	}
	return c.UpdateGenericModel(ctx, session, m)
}

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

func (c Client) SearchCategoriesByName(ctx context.Context, session *Session, name string, displayName string) ([]Category, error) {
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
