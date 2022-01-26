package model

import "github.com/vshn/appuio-odoo-adapter/odoo"

// Odoo is the developer-friendly odoo.Client with strongly-typed models.
type Odoo struct {
	session *odoo.Session
}

// NewOdoo creates a new Odoo client.
func NewOdoo(session *odoo.Session) *Odoo {
	return &Odoo{
		session: session,
	}
}
