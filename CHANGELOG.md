# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.2] - 2026-07-02

### Fixed
- Serialized pdfcpu's one-time global configuration initialization behind a
  `sync.Once`. pdfcpu lazily initializes a process-global default config on the
  first `NewDefaultConfiguration` call, and that init is not concurrency-safe;
  concurrent first callers (e.g. a server processing parallel PDF requests)
  raced on it. Embedding and extraction now go through a guarded helper.

## [0.1.1] - 2026-07-02

### Security
- Extraction now caps the decoded size of embedded files (64 MiB per attachment)
  and decodes only the attachment actually selected, one at a time, instead of
  eagerly decoding every attachment. Prevents decompression-bomb PDFs (Flate
  expands up to ~1000:1) from exhausting memory. Oversized attachments fail
  extraction with `filter.ErrDecodeLimitExceeded` when they carry the Factur-X
  name and are skipped in the content-sniffing fallback.
- The PDF/A detection during embedding decodes the XMP metadata stream with a
  16 MiB cap (same decompression-bomb guard).
- Bumped `golang.org/x/image` to v0.43.0 (GO-2026-5061: panic in the WebP
  decoder, reachable via pdfcpu).

### Changed
- Bumped the `go-xinvoice` dependency to v0.1.3, which bounds decimal length to
  prevent a `big.Rat` denial-of-service during validation (affects
  `ExtractInvoice` / any parsed invoice).

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
