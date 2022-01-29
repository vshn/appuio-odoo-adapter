package main

import "github.com/urfave/cli/v2"

const defaultTextForRequiredFlags = "<required>"

func newOdooURLFlag(destination *string) *cli.StringFlag {
	return &cli.StringFlag{Name: "odoo-url", Usage: "Odoo ERP URL in the form of https://username:password@host[:port]/db-name",
		EnvVars: envVars("ODOO_URL"), Destination: destination, Required: true, DefaultText: defaultTextForRequiredFlags}
}

func newDatabaseURLFlag(destination *string) *cli.StringFlag {
	return &cli.StringFlag{Name: "db-url", Usage: "PostgreSQL URL in the form of postgres://username:password@host[:port]/db-name",
		EnvVars: envVars("DB_URL"), Destination: destination, Required: true, DefaultText: defaultTextForRequiredFlags}
}
