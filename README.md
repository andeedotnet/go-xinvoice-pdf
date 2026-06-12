# go-xinvoice-pdf

A pure-Go companion to [`go-xinvoice`](../go-xinvoice) for **hybrid ZUGFeRD / Factur-X**
invoices: embed the XRechnung invoice XML into a PDF, and extract it back out.

A hybrid invoice is a **PDF/A-3** document that carries the machine-readable invoice as an
embedded *associated file*. This module:

1. **PDF → XML** — extract the embedded invoice XML from a hybrid PDF so it can be parsed,
   converted or validated with `go-xinvoice`.
2. **XML → PDF** — embed an XRechnung **CII** invoice XML into a PDF you provide, wiring up the
   PDF/A-3 Associated Files (`/AF`, `/AFRelationship Alternative`, `/Subtype text/xml`) and the
   Factur-X XMP metadata.

The PDF container is handled by [pdfcpu](https://github.com/pdfcpu/pdfcpu) (pure Go, no CGo).

> Status: early. See [`CHANGELOG.md`](./CHANGELOG.md).

## Install

```sh
go get github.com/andeedotnet/go-xinvoice-pdf
```

## Usage

```go
import (
    xinvoicepdf "github.com/andeedotnet/go-xinvoice-pdf"
    xinvoice "github.com/andeedotnet/go-xinvoice"
)

// Extract: hybrid PDF -> XML -> go-xinvoice model
xml, name, err := xinvoicepdf.ExtractXML(pdfBytes)
inv, err := xinvoicepdf.ExtractInvoice(pdfBytes) // extract + xinvoice.ParseXML

// Embed: CII XML into an existing PDF -> hybrid Factur-X PDF
out, warnings, err := xinvoicepdf.Embed(pdfBytes, ciiXML, nil) // nil = XRECHNUNG profile

// Embed from a go-xinvoice model (serializes to CII automatically)
out, warnings, err = xinvoicepdf.EmbedInvoice(pdfBytes, inv, &xinvoicepdf.Options{
    Profile: xinvoicepdf.ProfileEN16931,
})
```

`warnings` is non-empty when the input PDF is not already PDF/A (see Limitations).

## CLI

The `xinvoice-pdf` command (under `cmd/xinvoice-pdf`) wraps the same operations.

```sh
go install github.com/andeedotnet/go-xinvoice-pdf/cmd/xinvoice-pdf@latest

# Extract the embedded XML from a hybrid PDF
xinvoice-pdf extract --in invoice.pdf --out invoice.xml

# Embed CII XML into a PDF (default profile XRECHNUNG)
xinvoice-pdf embed --pdf invoice.pdf --in invoice.xml --profile xrechnung --out hybrid.pdf
```

The XML input to `embed` may be CII XML or model JSON (auto-detected via `go-xinvoice` and
serialized to CII).

## Profiles

The embedded syntax is always CII. The `--profile` flag selects the Factur-X / ZUGFeRD
conformance level written to the XMP metadata:

| Profile     | XMP ConformanceLevel | Notes                                   |
|-------------|----------------------|-----------------------------------------|
| `xrechnung` | `XRECHNUNG`          | default; the German XRechnung CIUS       |
| `en16931`   | `EN 16931`           | the generic EN 16931 (COMFORT) profile   |

## Limitations

- **PDF/A.** This module embeds the invoice and writes correct Factur-X Associated Files and
  XMP metadata, but it does **not** convert an arbitrary PDF into PDF/A. For a fully
  PDF/A-3-valid hybrid, supply a PDF that is already PDF/A (or run the output through a PDF/A
  converter). When the input is not PDF/A, `Embed` returns a warning and still writes a valid
  Factur-X hybrid.
- **CII only.** The hybrid format embeds the UN/CEFACT CII syntax; UBL is not embedded.
- For authoritative validation of the produced file, cross-check with veraPDF (PDF/A-3b) and
  the KoSIT / Mustang ZUGFeRD validator.

## License & attribution

Licensed under the MIT License — see [`LICENSE`](./LICENSE) and [`NOTICE`](./NOTICE).
Depends on pdfcpu (Apache-2.0) and go-xinvoice (MIT).
