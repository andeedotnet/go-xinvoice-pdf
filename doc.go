// Package xinvoicepdf embeds and extracts the XRechnung invoice XML into and out
// of hybrid ZUGFeRD / Factur-X PDFs (a PDF/A-3 document carrying the invoice XML
// as an associated file).
//
// It builds on the sibling module go-xinvoice for all XML work and on pdfcpu for
// the PDF container. The public surface is small:
//
//   - [ExtractXML] returns the embedded invoice XML from a hybrid PDF, and
//     [ExtractInvoice] parses it into a go-xinvoice [Invoice].
//   - [Embed] inserts a CII invoice XML into a user-provided PDF, and
//     [EmbedInvoice] serializes an [Invoice] to CII first.
//
// The embedded syntax is always UN/CEFACT CII (the syntax the ZUGFeRD/Factur-X
// hybrid format uses); the written XMP conformance level is chosen by [Profile]
// (XRECHNUNG by default, or EN 16931).
//
// Embedding wires up the PDF/A-3 Associated Files (catalog /AF, the filespec
// /AFRelationship "Alternative" and the stream /Subtype "text/xml") and the
// Factur-X XMP metadata, but does not convert an arbitrary input PDF into PDF/A:
// full PDF/A-3 validity requires the input to already be (close to) PDF/A. When
// the input is not PDF/A, [Embed] still produces a valid Factur-X hybrid and
// returns a warning.
package xinvoicepdf
