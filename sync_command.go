package main

import (
	"fmt"
	"net/url"

	"github.com/urfave/cli/v2"
)

type syncCommand struct {
	OdooURL      string
	OdooUsername string
	OdooPassword string
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
			newOdooUsernameFlag(&command.OdooUsername),
			newOdooPasswordFlag(&command.OdooPassword),
		},
	}
}

func (c *syncCommand) validate(context *cli.Context) error {
	_ = LogMetadata(context)
	log := AppLogger(context).WithName(syncCommandName)
	log.V(1).Info("validating config")
	// The `Required` property in the StringFlag above already checks if it's non-empty.
	if _, err := url.Parse(c.OdooURL); err != nil {
		return fmt.Errorf("could not parse %q as URL value for flag %s: %s", c.OdooURL, newOdooURLFlag(nil).Name, err)
	}
	return nil
}

func (c *syncCommand) execute(context *cli.Context) error {
	log := AppLogger(context).WithName(syncCommandName)
	log.Info("Hello from sync command!", "config", c)
	return nil
}

func (c *syncCommand) shutdown(context *cli.Context) error {
	log := AppLogger(context).WithName(syncCommandName)
	log.Info("Shutting down " + syncCommandName)
	return nil
}
