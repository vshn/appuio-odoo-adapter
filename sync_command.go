package main

import (
	"context"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/go-logr/logr"
	"github.com/urfave/cli/v2"
	"github.com/vshn/appuio-odoo-adapter/odoo"
	"github.com/vshn/appuio-odoo-adapter/odoo/model"
	"github.com/vshn/appuio-odoo-adapter/sync"
)

type syncCommand struct {
	OdooURL string
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
		},
	}
}

func (c *syncCommand) execute(context *cli.Context) error {
	_ = LogMetadata(context)
	log := AppLogger(context).WithName(syncCommandName)

	client, err := odoo.NewClient(c.OdooURL)
	if err != nil {
		return err
	}
	client.UseDebugLogger(context.Bool("debug"))
	log.V(1).Info("Logging in to Odoo...", "url", c.OdooURL, "db", client.DBName())

	odooCtx := logr.NewContext(context.Context, log)

	session, err := client.Login(odooCtx)
	if err != nil {
		return err
	}
	log.Info("Login succeeded", "uid", session.UID)

	// Demo Odoo API
	o := model.NewOdoo(session)
	c.demonstrateOdooAPI(odooCtx, o, log)

	log.Info("About to demonstrate a InvoiceCategoryReconciler")
	// Demo with Faked Reporting category
	rc := sync.NewInvoiceCategoryReconciler(o)
	cat := &db.Category{Source: "zone:namespace"}
	log.Info("Reconciling category", "category", cat)
	_, err = rc.Reconcile(odooCtx, cat)
	if err != nil {
		return err
	}
	log.Info("Reconciled category", "category", cat)
	return nil
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
