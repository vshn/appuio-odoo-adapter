package model

import "github.com/vshn/appuio-odoo-adapter/odoo"

// Odoo is the developer-friendly odoo.Client with strongly-typed models.
type Odoo struct {
	querier odoo.QueryExecutor
}

// NewOdoo creates a new Odoo client.
func NewOdoo(querier odoo.QueryExecutor) *Odoo {
	return &Odoo{
		querier: querier,
	}
}
