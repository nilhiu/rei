package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/nilhiu/rei/rasm"
	"github.com/urfave/cli/v3"
)

func main() {
	greenBold := color.New(color.FgGreen, color.Bold)

	cli.RootCommandHelpTemplate = greenBold.Sprint(
		"Usage:",
	) + " {{.Name}} {{if .VisibleFlags}}[options]{{end}} filename\n\n" +
		greenBold.Sprint(
			"Options:",
		) + "\n\t{{range .VisibleFlags}}{{.}}\n\t{{end}}\n"

	cmd := &cli.Command{
		Version: "0.1.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "",
				Usage:   "specifies the output `FILE` (defaults to {input_file}.bin)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			var input string
			if cmd.NArg() > 0 {
				input = cmd.Args().Get(0)
			} else {
				printErr("The Instructioneer wasn't given a file to assemble.")
				cli.ShowAppHelpAndExit(cmd, 2)
			}

			output := cmd.String("output")
			if output == "" {
				output = input + ".bin"
			}

			ok := assembleFileTo(input, output)
			if !ok {
				os.Exit(1)
			}

			printInfo("\"" + input + "\" was assembled to \"" + output + "\"")

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func assembleFileTo(input string, output string) bool {
	fin, err := os.Open(input)
	if err != nil {
		panic(err)
	}
	defer fin.Close()

	fout, err := os.Create(output)
	if err != nil {
		panic(err)
	}
	defer fout.Close()

	cg := rasm.NewCodeGen(fin)
	for {
		bs, _, err := cg.Next()
		if err != nil {
			printErr(err.Error())
			return false
		}

		if bs == nil {
			return true
		}

		_, err = fout.Write(bs)
		if err != nil {
			panic(err)
		}
	}
}

func printErr(msg string) {
	fmt.Fprintln(os.Stderr, color.New(color.FgRed, color.Bold).Sprint("[err]:"), msg)
}

func printInfo(msg string) {
	fmt.Fprintln(os.Stdout, color.New(color.FgBlue, color.Bold).Sprint("[info]:"), msg)
}
