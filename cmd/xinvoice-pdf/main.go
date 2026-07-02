// Command xinvoice-pdf embeds and extracts the XRechnung invoice XML into/from
// hybrid ZUGFeRD / Factur-X PDFs.
//
// Usage:
//
//	xinvoice-pdf embed   --pdf FILE [--in XML|JSON] [--out FILE] [--profile xrechnung|en16931]
//	xinvoice-pdf extract [--in PDF] [--out FILE]
//	xinvoice-pdf version
//
// `embed` attaches the invoice (read from --in or stdin, as CII XML, UBL XML or
// model JSON — always embedded as CII) to the PDF given by --pdf. `extract`
// writes the embedded XML to --out or stdout and exits non-zero when none is found.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	xinvoice "github.com/andeedotnet/go-xinvoice"
	xinvoicepdf "github.com/andeedotnet/go-xinvoice-pdf"
)

// version is the module version, kept in sync with the v0.1.1 git tag.
const version = "0.1.1"

func main() {
	if len(os.Args) < 2 {
		usage(os.Stderr)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "embed":
		os.Exit(cmdEmbed(os.Args[2:]))
	case "extract":
		os.Exit(cmdExtract(os.Args[2:]))
	case "-h", "--help", "help":
		usage(os.Stdout)
		os.Exit(0)
	case "version", "-v", "--version":
		fmt.Println("xinvoice-pdf", version)
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "xinvoice-pdf: unknown command %q\n\n", os.Args[1])
		usage(os.Stderr)
		os.Exit(2)
	}
}

func usage(w io.Writer) {
	fmt.Fprint(w, `xinvoice-pdf — embed and extract the XRechnung XML in hybrid ZUGFeRD/Factur-X PDFs

Commands:
  embed     Embed an invoice into a PDF as a Factur-X associated file
            --pdf FILE          base PDF to embed into (required)
            --in  FILE          invoice: CII/UBL XML or model JSON (default stdin)
            --out FILE          output hybrid PDF (default stdout)
            --profile NAME      xrechnung|en16931 (default xrechnung)

  extract   Extract the embedded invoice XML from a hybrid PDF
            --in  FILE          input PDF (default stdin)
            --out FILE          output XML (default stdout)
            exit status 1 when no embedded XML is found

  version   Print the xinvoice-pdf version

The invoice is always embedded as CII, whatever the input syntax.
`)
}

func cmdEmbed(args []string) int {
	fs := flag.NewFlagSet("embed", flag.ContinueOnError)
	pdf := fs.String("pdf", "", "base PDF to embed into (required)")
	in := fs.String("in", "-", "invoice file: CII/UBL XML or model JSON (- for stdin)")
	out := fs.String("out", "-", "output file (- for stdout)")
	profile := fs.String("profile", "xrechnung", "conformance profile: xrechnung|en16931")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *pdf == "" {
		fmt.Fprintln(os.Stderr, "xinvoice-pdf: --pdf is required")
		return 2
	}
	prof, err := xinvoicepdf.ParseProfile(*profile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	pdfBytes, err := os.ReadFile(*pdf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "xinvoice-pdf:", err)
		return 1
	}
	invData, err := readInput(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "xinvoice-pdf:", err)
		return 1
	}
	cii, err := toCII(invData)
	if err != nil {
		fmt.Fprintln(os.Stderr, "xinvoice-pdf:", err)
		return 1
	}

	hybrid, warnings, err := xinvoicepdf.Embed(pdfBytes, cii, &xinvoicepdf.Options{Profile: prof})
	if err != nil {
		fmt.Fprintln(os.Stderr, "xinvoice-pdf:", err)
		return 1
	}
	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, "xinvoice-pdf: warning:", w)
	}
	if err := writeOutput(*out, hybrid); err != nil {
		fmt.Fprintln(os.Stderr, "xinvoice-pdf:", err)
		return 1
	}
	return 0
}

func cmdExtract(args []string) int {
	fs := flag.NewFlagSet("extract", flag.ContinueOnError)
	in := fs.String("in", "-", "input PDF (- for stdin)")
	out := fs.String("out", "-", "output XML file (- for stdout)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	pdfBytes, err := readInput(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "xinvoice-pdf:", err)
		return 1
	}
	xml, _, err := xinvoicepdf.ExtractXML(pdfBytes)
	if err != nil {
		if errors.Is(err, xinvoicepdf.ErrNoEmbeddedXML) {
			fmt.Fprintln(os.Stderr, "xinvoice-pdf: no embedded invoice XML found")
			return 1
		}
		fmt.Fprintln(os.Stderr, "xinvoice-pdf:", err)
		return 1
	}
	if err := writeOutput(*out, xml); err != nil {
		fmt.Fprintln(os.Stderr, "xinvoice-pdf:", err)
		return 1
	}
	return 0
}

// toCII normalizes an invoice (CII XML, UBL XML or model JSON) to CII XML bytes.
// CII XML is returned verbatim; UBL and JSON are routed through the model.
func toCII(data []byte) ([]byte, error) {
	if t := bytes.TrimSpace(data); len(t) > 0 && t[0] == '<' {
		if bytes.Contains(t, []byte("CrossIndustryInvoice")) {
			return data, nil // already CII
		}
		inv, err := xinvoice.ParseXML(data)
		if err != nil {
			return nil, err
		}
		return inv.ToXML(xinvoice.CII)
	}
	inv, err := xinvoice.FromJSON(data)
	if err != nil {
		return nil, err
	}
	return inv.ToXML(xinvoice.CII)
}

func readInput(name string) ([]byte, error) {
	if name == "-" || name == "" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(name)
}

// writeOutput writes data verbatim (no trailing newline) so binary PDF output is
// not corrupted on stdout.
func writeOutput(name string, data []byte) error {
	if name == "-" || name == "" {
		_, err := os.Stdout.Write(data)
		return err
	}
	return os.WriteFile(name, data, 0o644)
}
