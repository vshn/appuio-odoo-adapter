package sync

import (
	"github.com/vshn/appuio-odoo-adapter/odoo/model"
)

// InvoiceCategoryReconciler synchronizes various reporting facts with Odoo API.
type InvoiceCategoryReconciler struct {
	odoo *model.Odoo
}
