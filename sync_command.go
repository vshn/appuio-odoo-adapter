package main

import (
	"fmt"
	"net/url"

	"github.com/urfave/cli/v2"
	"github.com/vshn/appuio-odoo-adapter/odoo"
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
	// The `Required` property in the StringFlag above already checks if it's non-empty.
	if _, err := url.Parse(c.OdooURL); err != nil {
		return fmt.Errorf("could not parse %q as URL value for flag %s: %s", c.OdooURL, newOdooURLFlag(nil).Name, err)
	}
	return nil
}

func (c *syncCommand) execute(context *cli.Context) error {
	_ = LogMetadata(context)
	log := AppLogger(context).WithName(syncCommandName)

	client := odoo.NewClient(c.OdooURL, c.OdooDB)
	log.V(1).Info("Logging in to Odoo...", "url", c.OdooURL, "db", c.OdooDB)
	session, err := client.Login(c.OdooUsername, c.OdooPassword)
	if err != nil {
		return err
	}
	log.Info("Login succeeded", "uid", session.UID)
	return nil
}

func (c *syncCommand) shutdown(context *cli.Context) error {
	log := AppLogger(context).WithName(syncCommandName)
	log.Info("Shutting down " + syncCommandName)
	return nil
}
