package main

import "github.com/urfave/cli/v2"

const defaultTextForRequiredFlags = "<required>"

func newOdooURLFlag(destination *string) *cli.StringFlag {
	return &cli.StringFlag{Name: "odoo-url", Usage: "Odoo ERP URL in the form of https://host/",
		EnvVars: envVars("ODOO_URL"), Destination: destination, Required: true, DefaultText: defaultTextForRequiredFlags}
}

func newOdooDBFlag(destination *string) *cli.StringFlag {
	return &cli.StringFlag{Name: "odoo-db", Usage: "Odoo Database name",
		EnvVars: envVars("ODOO_DB"), Destination: destination, Required: true, DefaultText: defaultTextForRequiredFlags}
}

func newOdooUsernameFlag(destination *string) *cli.StringFlag {
	return &cli.StringFlag{Name: "odoo-username", Usage: "Odoo ERP username for authentication",
		EnvVars: envVars("ODOO_USERNAME"), Destination: destination, Required: true, DefaultText: defaultTextForRequiredFlags}
}

func newOdooPasswordFlag(destination *string) *cli.StringFlag {
	return &cli.StringFlag{Name: "odoo-password", Usage: "Odoo ERP password for authentication",
		EnvVars: envVars("ODOO_PASSWORD"), Destination: destination, Required: true, DefaultText: defaultTextForRequiredFlags}
}
