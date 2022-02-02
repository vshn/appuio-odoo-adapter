package main

import (
	"os"
	"path/filepath"

	"github.com/appuio/appuio-cloud-reporting/pkg/categories"
	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/go-logr/logr"
	"github.com/urfave/cli/v2"
	"github.com/vshn/appuio-odoo-adapter/odoo"
	"github.com/vshn/appuio-odoo-adapter/odoo/model"
	"github.com/vshn/appuio-odoo-adapter/sync"
	"gopkg.in/yaml.v3"
)

type syncCommand struct {
	OdooURL     string
	DatabaseURL string

	ZoneNameFile string
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
			&cli.StringFlag{Name: "zone-name-file", Usage: "Path to a file with zone name mappings.",
				EnvVars: envVars("ZONE_NAME_FILE"), Destination: &command.ZoneNameFile, Value: "zone-names.yaml", Required: false},
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

	log.V(1).Info("Opening database connection...")
	rdb, err := db.Openx(c.DatabaseURL)
	if err != nil {
		return err
	}
	defer rdb.Close()

	log.V(1).Info("loading zone name mappings...")
	mapper, err := c.zoneNameMapper()
	if err != nil {
		return err
	}

	o := model.NewOdoo(session)
	rc := sync.NewInvoiceCategoryReconciler(o)
	rc.ZoneNameMapper = mapper

	err = categories.Reconcile(odooCtx, rdb, rc)
	return err
}

func (c *syncCommand) zoneNameMapper() (sync.ZoneNameMapper, error) {
	if c.ZoneNameFile == "" {
		return nil, nil
	}

	raw, err := os.ReadFile(filepath.Join(".", c.ZoneNameFile))
	if err != nil {
		return nil, err
	}
	var mappings map[string]string
	err = yaml.Unmarshal([]byte(raw), &mappings)
	if err != nil {
		return nil, err
	}

	return &sync.StaticZoneMapper{Map: mappings}, nil
}
