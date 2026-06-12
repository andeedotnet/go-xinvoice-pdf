# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-06-12

### Added
- Initial module: embed and extract the XRechnung invoice XML into/from hybrid
  ZUGFeRD / Factur-X PDFs.
- `Embed` / `EmbedInvoice` — attach the CII XML as `factur-x.xml` to a user-provided
  PDF, wiring up the PDF/A-3 Associated Files (`/AF`, `/AFRelationship Alternative`,
  `/Subtype text/xml`) and the Factur-X XMP metadata.
- `ExtractXML` / `ExtractInvoice` — locate and return the embedded invoice XML.
- Conformance profiles `XRECHNUNG` (default) and `EN 16931`, selectable via the
  `--profile` flag.
- `cmd/xinvoice-pdf` CLI with `extract`, `embed` and `version` subcommands.
