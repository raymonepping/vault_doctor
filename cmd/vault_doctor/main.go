package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/raymonepping/vault_doctor/internal/medic"
)

var Version = "v" // overridden by -ldflags

func main() {
	if len(os.Args) == 1 {
		fmt.Print(medic.Doc(Version))
		return
	}

	switch os.Args[1] {
	case "-h", "--help", "help":
		fmt.Print(medic.Doc(Version))
		return
	case "-V", "--version":
		fmt.Println(strings.TrimPrefix(Version, "v"))
		return
	case "medic":
		runMedicCmd()
	default:
		fmt.Print(medic.Doc(Version))
	}
}

func runMedicCmd() {
	fs := flag.NewFlagSet("medic", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "Output JSON")
	quiet := fs.Bool("quiet", false, "Quiet mode")
	noColor := fs.Bool("no-color", false, "Disable colors")
	_ = fs.Parse(os.Args[2:])

	opt := medic.Options{
		Version: Version,
		Quiet:   *quiet,
		JSON:    *jsonOut,
		NoColor: *noColor,
	}

	code := medic.Run(opt)
	os.Exit(code)
}
