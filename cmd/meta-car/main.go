package main

import (
	"fmt"
	"github.com/multiformats/go-multicodec"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	app := &cli.App{
		Name:  "meta-car",
		Usage: "Utility for working with car files",
		Commands: []*cli.Command{
			{
				Name:   "compile",
				Usage:  "compile a car file from a debug patch",
				Action: CompileCar,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:      "output",
						Aliases:   []string{"o", "f"},
						Usage:     "The file to write to",
						TakesFile: true,
					},
				},
			},
			{
				Name:    "create",
				Usage:   "Create a car file",
				Aliases: []string{"c"},
				Action:  CreateCar,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:      "file",
						Aliases:   []string{"f", "output", "o"},
						Usage:     "The car file to write to",
						TakesFile: true,
					},
					&cli.IntFlag{
						Name:  "version",
						Value: 2,
						Usage: "Write output as a v1 or v2 format car",
					},
				},
			},
			{
				Name:    "append",
				Usage:   "Append files to a car file",
				Aliases: []string{"c"},
				Action:  CreateCar,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:      "file",
						Aliases:   []string{"f", "output", "o"},
						Usage:     "The car file to write to",
						TakesFile: true,
					},
					&cli.IntFlag{
						Name:  "version",
						Value: 2,
						Usage: "Write output as a v1 or v2 format car",
					},
				},
			},
			{
				Name:   "debug",
				Usage:  "debug a car file",
				Action: DebugCar,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:      "output",
						Aliases:   []string{"o", "f"},
						Usage:     "The file to write to",
						TakesFile: true,
					},
				},
			},
			{
				Name:   "detach-index",
				Usage:  "Detach an index to a detached file",
				Action: DetachCar,
				Subcommands: []*cli.Command{{
					Name:   "list",
					Usage:  "List a detached index",
					Action: DetachCarList,
				}},
			},
			{
				Name:    "extract",
				Aliases: []string{"x"},
				Usage:   "Extract the contents of a car when the car encodes UnixFS data",
				Action:  ExtractCar,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:      "file",
						Aliases:   []string{"f"},
						Usage:     "The car file to extract from",
						Required:  true,
						TakesFile: true,
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Include verbose information about extracted contents",
					},
				},
			},
			{
				Name:    "filter",
				Aliases: []string{"f"},
				Usage:   "Filter the CIDs in a car",
				Action:  FilterCar,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:      "cid-file",
						Usage:     "A file to read CIDs from",
						TakesFile: true,
					},
					&cli.BoolFlag{
						Name:  "append",
						Usage: "Append cids to an existing output file",
					},
				},
			},
			{
				Name:    "get-block",
				Aliases: []string{"gb"},
				Usage:   "Get a block out of a car",
				Action:  GetCarBlock,
			},
			{
				Name:    "get-dag",
				Aliases: []string{"gd"},
				Usage:   "Get a dag out of a car",
				Action:  GetCarDag,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "selector",
						Aliases: []string{"s"},
						Usage:   "A selector over the dag",
					},
					&cli.BoolFlag{
						Name:  "strict",
						Usage: "Fail if the selector finds links to blocks not in the original car",
					},
					&cli.IntFlag{
						Name:  "version",
						Value: 2,
						Usage: "Write output as a v1 or v2 format car",
					},
				},
			},
			{
				Name:    "index",
				Aliases: []string{"i"},
				Usage:   "write out the car with an index",
				Action:  IndexCar,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "codec",
						Aliases: []string{"c"},
						Usage:   "The type of index to write",
						Value:   multicodec.CarMultihashIndexSorted.String(),
					},
					&cli.IntFlag{
						Name:  "version",
						Value: 2,
						Usage: "Write output as a v1 or v2 format car",
					},
				},
			},
			{
				Name:   "inspect",
				Usage:  "verifies a car and prints a basic report about its contents",
				Action: InspectCar,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "full",
						Value: false,
						Usage: "Check that the block data hash digests match the CIDs",
					},
				},
			},
			{
				Name:    "list",
				Aliases: []string{"l", "ls"},
				Usage:   "List the CIDs in a car",
				Action:  ListCar,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Include verbose information about contained blocks",
					},
					&cli.BoolFlag{
						Name:  "unixfs",
						Usage: "List unixfs filesystem from the root of the car",
					},
					&cli.BoolFlag{
						Name:  "links",
						Usage: "List links from the root of the car",
					},
				},
			},
			{
				Name:   "root",
				Usage:  "Get the root CID of a car",
				Action: CarRoot,
			},
			{
				Name:    "verify",
				Aliases: []string{"v"},
				Usage:   "Verify a CAR is wellformed",
				Action:  VerifyCar,
			},
			{
				Name:   "test",
				Usage:  "test build a car from files",
				Action: CreateCarFileTest,
			},
			{
				Name:  "build",
				Usage: "Generate CAR files of the specified size",
				Flags: []cli.Flag{
					&cli.Uint64Flag{
						Name:  "slice-size",
						Value: 17179869184, // 16G
						Usage: "specify chunk piece size",
					},
					&cli.UintFlag{
						Name:  "parallel",
						Value: 2,
						Usage: "specify how many number of goroutines runs when generate file node",
					},
					&cli.StringFlag{
						Name:  "graph-name",
						Value: "meta",
						Usage: "specify graph name",
					},
					&cli.StringFlag{
						Name:     "car-dir",
						Required: true,
						Usage:    "specify output CAR directory",
					},
					&cli.StringFlag{
						Name:     "uuid",
						Required: true,
						Usage:    "Add uuid to filename suffix",
					},
					&cli.StringFlag{
						Name:  "parent-path",
						Value: "",
						Usage: "specify graph parent path",
					},
					&cli.BoolFlag{
						Name:  "save-manifest",
						Value: true,
						Usage: "create a mainfest.csv in car-dir to save mapping of data-cids and slice names",
					},
				},
				Action: CarBuild,
			},
			{
				Name:  "restore",
				Usage: "Restore files from CAR files",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "car-path",
						Required: true,
						Usage:    "specify source car path, directory or file",
					},
					&cli.StringFlag{
						Name:     "output-dir",
						Required: true,
						Usage:    "specify output directory",
					},
					&cli.IntFlag{
						Name:  "parallel",
						Value: 2,
						Usage: "specify how many number of goroutines runs when generate file node",
					},
				},
				Action: Restore,
			},
			{
				Name:  "restore-ex",
				Usage: "Restore files from CAR files",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "car-path",
						Required: true,
						Usage:    "specify source car path, directory or file",
					},
					&cli.StringFlag{
						Name:     "output-dir",
						Required: true,
						Usage:    "specify output directory",
					},
					&cli.IntFlag{
						Name:  "parallel",
						Value: 2,
						Usage: "specify how many number of goroutines runs when generate file node",
					},
				},
				Action: RestoreEx,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}
