package sync

import (
	"context"
	"os"

	"github.com/appuio/appuio-cloud-reporting/pkg/erp"
	"github.com/vshn/appuio-odoo-adapter/odoo"
	"github.com/vshn/appuio-odoo-adapter/odoo/model"
)

var adapterInstance = &OdooAdapter{}

// OdooAdapter implements erp.Adapter.
type OdooAdapter struct {
	odoo *model.Odoo
}

func init() {
	erp.Register(adapterInstance)
}

// Initialize implements erp.Adapter.
func (o *OdooAdapter) Initialize(ctx context.Context) error {
	if o.odoo != nil {
		return nil
	}
	return o.initializeSession(ctx)
}

// NewCategoryReconciler implements erp.Adapter.
func (o *OdooAdapter) NewCategoryReconciler() erp.CategoryReconciler {
	return NewInvoiceCategoryReconciler(o.odoo)
}

func (o *OdooAdapter) initializeSession(ctx context.Context) error {
	odooUrl := os.Getenv("OA_ODOO_URL")
	client, err := odoo.NewClient(odooUrl, "db")
	if err != nil {
		return err
	}
	session, err := client.Login(ctx, "", "")
	if err != nil {
		return err
	}
	o.odoo = model.NewOdoo(session)
	return nil
}
