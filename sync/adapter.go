package sync

import (
	"context"
	"os"

	"github.com/appuio/appuio-cloud-reporting/pkg/erp"
	"github.com/vshn/appuio-odoo-adapter/odoo"
	"github.com/vshn/appuio-odoo-adapter/odoo/model"
)

// EnvVarPrefix is the global prefix for all environment variables to configure the adapter.
var EnvVarPrefix = "OA_"

var adapterInstance = &OdooAdapter{}

// OdooAdapter implements erp.Adapter.
type OdooAdapter struct{}

// OdooDriver implements erp.Driver.
type OdooDriver struct {
	odoo *model.Odoo
}

func init() {
	erp.Register(adapterInstance)
}

// Open implements erp.Adapter.
func (o *OdooAdapter) Open(ctx context.Context) (erp.Driver, error) {
	driver := &OdooDriver{}
	return driver, driver.initializeSession(ctx)
}

func (o *OdooDriver) initializeSession(ctx context.Context) error {
	odooURL := getEnv("ODOO_URL")
	useDebug := getEnv("DEBUG") == "true"
	session, err := odoo.Open(ctx, odooURL, odoo.ClientOptions{UseDebugLogger: useDebug})
	if err != nil {
		return err
	}
	o.odoo = model.NewOdoo(session)
	return nil
}

// NewCategoryReconciler implements erp.Driver.
func (o *OdooDriver) NewCategoryReconciler() erp.CategoryReconciler {
	return NewInvoiceCategoryReconciler(o.odoo)
}

// Close implements erp.Driver.
func (o *OdooDriver) Close(ctx context.Context) error {
	return nil
}

func getEnv(name string) string {
	return os.Getenv(EnvVarPrefix + name)
}
