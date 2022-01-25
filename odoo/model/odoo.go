package model

import "github.com/vshn/appuio-odoo-adapter/odoo"

type Odoo struct {
	client  *odoo.Client
	session *odoo.Session
}

// NewOdoo creates a new Odoo client.
func NewOdoo(session *odoo.Session, client *odoo.Client) *Odoo {
	return &Odoo{
		client:  client,
		session: session,
	}
}
