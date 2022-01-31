package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	reportinvoice "github.com/appuio/appuio-cloud-reporting/pkg/invoice"
	"github.com/go-logr/logr"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/vshn/appuio-odoo-adapter/invoice"
	"github.com/vshn/appuio-odoo-adapter/odoo"
	"github.com/vshn/appuio-odoo-adapter/odoo/model"
)

//go:embed invoice-defaults.yaml
var invoiceDefaultsYAML string

type invoiceCommand struct {
	OdooURL     string
	DatabaseURL string
	Year        int
	Month       time.Month

	InvoiceDefaultsPath string
}

var invoiceCommandName = "invoice"

func newinvoiceCommand() *cli.Command {
	command := &invoiceCommand{}
	return &cli.Command{
		Name:   invoiceCommandName,
		Usage:  "Create Odoo invoices from APPUiO Cloud",
		Action: command.execute,
		Flags: []cli.Flag{
			newOdooURLFlag(&command.OdooURL),
			newDatabaseURLFlag(&command.DatabaseURL),

			&cli.IntFlag{Name: "year", Usage: "Year to generate the report for.",
				EnvVars: envVars("YEAR"), Destination: &command.Year, Required: true},
			&cli.IntFlag{Name: "month", Usage: "Month to generate the report for.",
				EnvVars: envVars("MONTH"), Destination: (*int)(&command.Month), Required: true},
			&cli.StringFlag{Name: "invoice-defaults-path", Usage: "Path to a file with invoice defaults.",
				EnvVars: envVars("INVOICE_DEFAULTS_PATH"), Destination: &command.InvoiceDefaultsPath, Required: false},
		},
	}
}

func (cmd *invoiceCommand) execute(context *cli.Context) error {
	ctx := context.Context
	_ = LogMetadata(context)
	log := AppLogger(context).WithName(invoiceCommandName)

	invDefault, invLineDefault, err := cmd.loadInvoiceDefaults()
	if err != nil {
		return fmt.Errorf("failed to load defaults: %w", err)
	}

	odooCtx := logr.NewContext(context.Context, log)
	log.V(1).Info("Logging in to Odoo...")
	session, err := odoo.Open(odooCtx, cmd.OdooURL, odoo.ClientOptions{UseDebugLogger: context.Bool("debug")})
	if err != nil {
		return err
	}
	log.Info("login succeeded", "uid", session.UID)

	o := model.NewOdoo(session)

	log.V(1).Info("Opening database connection...")
	rdb, err := db.Openx(cmd.DatabaseURL)
	if err != nil {
		return err
	}
	defer rdb.Close()

	log.V(1).Info("Begin transaction")
	tx, err := rdb.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	invoices, err := reportinvoice.Generate(ctx, tx, cmd.Year, cmd.Month)
	if err != nil {
		return err
	}

	for _, inv := range invoices {
		id, err := invoice.CreateInvoice(ctx, o, inv,
			invoice.WithInvoiceDefaults(invDefault),
			invoice.WithInvoiceLineDefaults(invLineDefault),
		)
		log.Info("Created invoice", "id", id)
		if err != nil {
			return fmt.Errorf("error creating invoice %+v: %w", inv, err)
		}
	}

	return nil
}

func (cmd *invoiceCommand) loadInvoiceDefaults() (model.Invoice, model.InvoiceLine, error) {
	type load struct {
		Invoice     model.Invoice     `yaml:"invoice"`
		InvoiceLine model.InvoiceLine `yaml:"invoice_line"`
	}

	raw := []byte(invoiceDefaultsYAML)
	if cmd.InvoiceDefaultsPath != "" {
		var err error
		raw, err = os.ReadFile(filepath.Join(".", cmd.InvoiceDefaultsPath))
		if err != nil {
			return model.Invoice{}, model.InvoiceLine{}, fmt.Errorf("error reading defaults file: %w", err)
		}
	}

	var out load
	err := yaml.Unmarshal([]byte(raw), &out)
	return out.Invoice, out.InvoiceLine, err
}
