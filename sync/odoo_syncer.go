package sync

import "github.com/vshn/appuio-odoo-adapter/odoo/model"

// OdooSyncer synchronizes various reporting facts with Odoo API.
type OdooSyncer struct {
	odoo *model.Odoo
}
