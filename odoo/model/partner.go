package model

import (
	"context"

	"github.com/vshn/appuio-odoo-adapter/odoo"
)

type Partner struct {
	ID   int    `json:"id,omitempty" yaml:"id,omitempty"`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

type PartnerList struct {
	Items []Partner `json:"records"`
}

// FetchPartnerByID searches for the partner by ID and returns the first entry in the result.
// If no result has been found, nil is returned without error.
func (o Odoo) FetchPartnerByID(ctx context.Context, id int) (*Partner, error) {
	result, err := o.searchPartners(ctx, []odoo.Filter{
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

func (o Odoo) searchPartners(ctx context.Context, domainFilters []odoo.Filter) ([]Partner, error) {
	result := &PartnerList{}
	err := o.querier.SearchGenericModel(ctx, odoo.SearchReadModel{
		Model:  "res.partner",
		Domain: domainFilters,
		Fields: []string{"name"},
	}, result)
	return result.Items, err
}
