package main

import (
	"log"
	"os"
	"time"

	"github.com/prgres/vid2asci/cmd/app"
	cli "github.com/urfave/cli/v2"
)

var (
	fDebug = "debug"
	fInput = "input"
)

func main() {
	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:  fDebug,
			Value: false,
			Usage: "debug",
		},
	}

	app := &cli.App{
		EnableBashCompletion: true,
		Name:                 "vid2asci",
		Usage:                "",
		Description:          "Run your favorite shows directly in your favorite term (with some laggy and fuzzy frames of course)",
		Flags:                flags,
		Compiled:             time.Now(),
		Authors: []*cli.Author{
			{
				Name: "M. Więcek",
			},
		},
		Before: func(c *cli.Context) error {
			return app.Before(c.Bool(fDebug))
		},
		Commands: []*cli.Command{
			{
				Name: "render",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     fInput,
						Aliases:  []string{"i"},
						Required: true,
						Usage:    "Video input path",
					},
				},
				Action: func(c *cli.Context) error {
					return app.Render(c.String(fInput))
				},
			},
			{
				Name: "play",
				Action: func(c *cli.Context) error {
					return app.Play()
				},
			},
			{
				Name: "start",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     fInput,
						Aliases:  []string{"i"},
						Required: true,
						Usage:    "Video input path",
					},
				},
				Action: func(c *cli.Context) error {
					if err := app.Render(c.String(fInput)); err != nil {
						return err
					}

					if err := app.Play(); err != nil {
						return err
					}

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
		return
	}
}
