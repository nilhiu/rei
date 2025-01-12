package main

import (
	"bytes"
	"context"
	"debug/elf"
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/nilhiu/rei/rasm"
	"github.com/nilhiu/rei/relf"
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
			&cli.BoolFlag{
				Name:  "binary",
				Value: false,
				Usage: "tells rei to only output machine code (no object file)",
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

			isBinOut := cmd.Bool("binary")
			output := cmd.String("output")
			if output == "" {
				output = input + ".bin"
			}

			var ok bool
			if isBinOut {
				ok = assembleBinary(input, output)
			} else {
				ok = assembleELF(input, output)
			}

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

func assembleBinary(input string, output string) bool {
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

// TODO: Find a clearer way to do this...
func assembleELF(input string, output string) bool {
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

	sectCode := map[string]*bytes.Buffer{}
	sectIndex := map[string]uint16{}

	cg := rasm.NewCodeGen(fin)
	for {
		bs, sect, err := cg.Next()
		if err != nil {
			printErr(err.Error())
			return false
		}

		if bs == nil {
			break
		}

		if buf, ok := sectCode[sect]; !ok {
			sectCode[sect] = bytes.NewBuffer(bs)
		} else {
			buf.Write(bs)
		}
	}

	w := relf.New(input, relf.Header64{
		Endian:  elf.ELFDATA2LSB,
		ABI:     elf.ELFOSABI_NONE,
		Machine: elf.EM_X86_64,
	}, fout)

	i := 1
	for k, buf := range sectCode {
		fmt.Println("section", k, "size", buf.Len())
		err := w.WriteSection(relf.Section64{
			Name:      k,
			Type:      elf.SHT_PROGBITS,
			Addralign: 16,
			Entsize:   0,
			Flags:     elf.SHF_EXECINSTR | elf.SHF_ALLOC,
			Code:      buf.Bytes(),
		})
		if err != nil {
			return false
		}

		sectIndex[k] = uint16(i)
		i++
	}

	for k, li := range cg.Labels() {
		err := w.WriteSymbol(relf.Symbol64{
			Name:  k,
			Type:  elf.STT_NOTYPE,
			Bind:  elf.STB_GLOBAL,
			Shndx: sectIndex[li.Section],
			Value: li.Offset,
		})
		if err != nil {
			return false
		}
	}

	if err := w.Flush(); err != nil {
		return false
	}

	return true
}

func printErr(msg string) {
	fmt.Fprintln(os.Stderr, color.New(color.FgRed, color.Bold).Sprint("[err]:"), msg)
}

func printInfo(msg string) {
	fmt.Fprintln(os.Stdout, color.New(color.FgBlue, color.Bold).Sprint("[info]:"), msg)
}
