package main

import (
	"fmt"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/urfave/cli/v2"
	"github.com/vshn/appuio-odoo-adapter/odoo"
	"github.com/vshn/appuio-odoo-adapter/odoo/model"
)

type syncCommand struct {
	OdooURL      string
	OdooUsername string
	OdooPassword string
	OdooDB       string
}

var syncCommandName = "sync"

func newSyncCommand() *cli.Command {
	command := &syncCommand{}
	return &cli.Command{
		Name:   syncCommandName,
		Usage:  "Sync Odoo entities from APPUiO Cloud",
		Before: command.validate,
		Action: command.execute,
		Flags: []cli.Flag{
			newOdooURLFlag(&command.OdooURL),
			newOdooDBFlag(&command.OdooDB),
			newOdooUsernameFlag(&command.OdooUsername),
			newOdooPasswordFlag(&command.OdooPassword),
		},
	}
}

func (c *syncCommand) validate(context *cli.Context) error {
	log := AppLogger(context).WithName(syncCommandName)
	log.V(1).Info("validating config")
	// The `Required` property in the StringFlag already checks if it's non-empty.
	if _, err := url.Parse(c.OdooURL); err != nil {
		return fmt.Errorf("could not parse %q as URL value for flag %s: %s", c.OdooURL, newOdooURLFlag(nil).Name, err)
	}
	return nil
}

func (c *syncCommand) execute(context *cli.Context) error {
	_ = LogMetadata(context)
	log := AppLogger(context).WithName(syncCommandName)

	client, err := odoo.NewClient(c.OdooURL, c.OdooDB)
	if err != nil {
		return err
	}
	client.UseDebugLogger(context.Bool("debug"))
	log.V(1).Info("Logging in to Odoo...", "url", c.OdooURL, "db", c.OdooDB)

	odooCtx := logr.NewContext(context.Context, log)

	session, err := client.Login(odooCtx, c.OdooUsername, c.OdooPassword)
	if err != nil {
		return err
	}
	log.Info("Login succeeded", "uid", session.UID)

	// Demo
	models := model.NewOdoo(session)
	createCategory := model.InvoiceCategory{
		Name:      "test-category-odoo-adapter",
		Sequence:  10,
		Separator: true,
		SubTotal:  true,
	}

	newCategory, err := models.CreateInvoiceCategory(odooCtx, createCategory)
	c.logIfErr(log, err)
	log.Info("Created new category", "category", newCategory)

	category, err := models.FetchInvoiceCategoryByID(odooCtx, newCategory.ID)
	log.Info("Fetched category", "category", category)

	newCategory.Sequence = 20
	err = models.UpdateInvoiceCategory(odooCtx, newCategory)
	c.logIfErr(log, err)
	log.Info("Updated category", "category", newCategory)

	list, err := models.SearchInvoiceCategoriesByName(odooCtx, "odoo-adapter")
	c.logIfErr(log, err)
	log.Info("Fetched list", "list", list)

	err = models.DeleteInvoiceCategory(odooCtx, newCategory)
	log.Info("Deleted new category", "category", newCategory)
	c.logIfErr(log, err)

	return nil
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
