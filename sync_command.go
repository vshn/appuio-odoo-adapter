package main

import (
	"context"

	"github.com/appuio/appuio-cloud-reporting/pkg/categories"
	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/go-logr/logr"
	"github.com/urfave/cli/v2"
	"github.com/vshn/appuio-odoo-adapter/odoo"
	"github.com/vshn/appuio-odoo-adapter/odoo/model"
	"github.com/vshn/appuio-odoo-adapter/sync"
)

type syncCommand struct {
	OdooURL     string
	DatabaseURL string
}

var syncCommandName = "sync"

func newSyncCommand() *cli.Command {
	command := &syncCommand{}
	return &cli.Command{
		Name:   syncCommandName,
		Usage:  "Sync Odoo entities from APPUiO Cloud",
		Action: command.execute,
		Flags: []cli.Flag{
			newOdooURLFlag(&command.OdooURL),
			newDatabaseURLFlag(&command.DatabaseURL),
		},
	}
}

func (c *syncCommand) execute(context *cli.Context) error {
	_ = LogMetadata(context)
	log := AppLogger(context).WithName(syncCommandName)

	odooCtx := logr.NewContext(context.Context, log)
	log.V(1).Info("Logging in to Odoo...")
	session, err := odoo.Open(odooCtx, c.OdooURL, odoo.ClientOptions{UseDebugLogger: context.Bool("debug")})
	if err != nil {
		return err
	}
	log.Info("login succeeded", "uid", session.UID)

	// Demo Odoo API
	o := model.NewOdoo(session)
	//c.demonstrateOdooAPI(odooCtx, o, log)

	rc := sync.NewInvoiceCategoryReconciler(o)

	log.V(1).Info("Opening database connection...")
	rdb, err := db.Openx(c.DatabaseURL)
	defer rdb.Close()

	err = categories.Reconcile(odooCtx, rdb, rc)
	return err
}

func (c *syncCommand) demonstrateOdooAPI(odooCtx context.Context, odoo *model.Odoo, log logr.Logger) {
	createCategory := model.InvoiceCategory{
		Name:      "test-category-odoo-adapter",
		Sequence:  10,
		Separator: true,
		SubTotal:  true,
	}

	newCategory, err := odoo.CreateInvoiceCategory(odooCtx, createCategory)
	c.logIfErr(log, err)
	log.Info("Created new category", "category", newCategory)

	category, err := odoo.FetchInvoiceCategoryByID(odooCtx, newCategory.ID)
	log.Info("Fetched category", "category", category)

	newCategory.Sequence = 20
	err = odoo.UpdateInvoiceCategory(odooCtx, newCategory)
	c.logIfErr(log, err)
	log.Info("Updated category", "category", newCategory)

	list, err := odoo.SearchInvoiceCategoriesByName(odooCtx, "odoo-adapter")
	c.logIfErr(log, err)
	log.Info("Fetched list", "list", list)

	err = odoo.DeleteInvoiceCategory(odooCtx, newCategory)
	log.Info("Deleted new category", "category", newCategory)
	c.logIfErr(log, err)
}

func (c *syncCommand) logIfErr(logger logr.Logger, err error) {
	if err != nil {
		logger.Error(err, "demo failed")
	}
}

func (c *syncCommand) shutdown(context *cli.Context) error {
	log := AppLogger(context).WithName(syncCommandName)
	log.Info("Shutting down " + syncCommandName)
	return nil
}
