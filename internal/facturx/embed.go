package facturx

import (
	"bytes"
	"fmt"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"

	"github.com/andeedotnet/go-xinvoice-pdf/internal/xmp"
)

// Options configures Embed.
type Options struct {
	// ConformanceLevel is written to the Factur-X XMP fx:ConformanceLevel field
	// (one of the xmp.Conformance* constants). Defaults to XRECHNUNG.
	ConformanceLevel string
	// AFRelationship is the filespec /AFRelationship value. Defaults to
	// DefaultAFRelationship ("Alternative").
	AFRelationship string
	// Description is the filespec /Desc. Defaults to "Factur-X invoice".
	Description string
	// ModTime is the embedded file's modification date. Defaults to time.Now.
	ModTime time.Time
}

func (o *Options) applyDefaults() {
	if o.ConformanceLevel == "" {
		o.ConformanceLevel = xmp.ConformanceXRechnung
	}
	if o.AFRelationship == "" {
		o.AFRelationship = DefaultAFRelationship
	}
	if o.Description == "" {
		o.Description = "Factur-X invoice"
	}
	if o.ModTime.IsZero() {
		o.ModTime = time.Now()
	}
}

// Embed inserts the CII invoice XML into pdf as a Factur-X associated file and
// returns the resulting hybrid PDF. warnings is non-empty when the input is not
// already PDF/A: the output is still a valid Factur-X hybrid, but full PDF/A-3
// validity requires a PDF/A input (pdfcpu does not convert to PDF/A).
func Embed(pdf, ciiXML []byte, opt Options) (out []byte, warnings []string, err error) {
	opt.applyDefaults()

	conf := newConfiguration()
	conf.ValidationMode = model.ValidationRelaxed
	// Write a classic cross-reference table and no object streams: simpler output,
	// broadly accepted by PDF/A validators, and keeps the associated-file dicts as
	// plaintext objects. (Object streams would force an xref stream.)
	conf.WriteObjectStream = false
	conf.WriteXRefStream = false

	ctx, err := api.ReadContext(bytes.NewReader(pdf), conf)
	if err != nil {
		return nil, nil, fmt.Errorf("xinvoice-pdf: read PDF: %w", err)
	}

	if !isPDFA(ctx) {
		warnings = append(warnings,
			"input PDF is not PDF/A; the output is a valid Factur-X hybrid but may not pass strict PDF/A-3 validation")
	}

	if err := attachInvoice(ctx.XRefTable, ciiXML, opt); err != nil {
		return nil, nil, fmt.Errorf("xinvoice-pdf: attach invoice: %w", err)
	}
	if err := setMetadata(ctx.XRefTable, xmp.Packet(opt.ConformanceLevel)); err != nil {
		return nil, nil, fmt.Errorf("xinvoice-pdf: set XMP metadata: %w", err)
	}

	var buf bytes.Buffer
	if err := api.WriteContext(ctx, &buf); err != nil {
		return nil, nil, fmt.Errorf("xinvoice-pdf: write PDF: %w", err)
	}
	return buf.Bytes(), warnings, nil
}

// maxDecodedMetadataBytes caps the decoded size of the XMP metadata stream
// inspected by isPDFA — the same decompression-bomb guard as for attachment
// extraction. Real XMP packets are a few KB.
const maxDecodedMetadataBytes = 16 << 20

// isPDFA reports whether the document already declares PDF/A conformance in its
// XMP metadata (looked up by the pdfaid:part marker).
func isPDFA(ctx *model.Context) bool {
	root, err := ctx.XRefTable.Catalog()
	if err != nil {
		return false
	}
	o, found := root.Find("Metadata")
	if !found {
		return false
	}
	sd, _, err := ctx.XRefTable.DereferenceStreamDict(o)
	if err != nil || sd == nil {
		return false
	}
	data := sd.Content
	if len(data) == 0 {
		if err := sd.DecodeWithLimit(maxDecodedMetadataBytes); err == nil {
			data = sd.Content
		}
	}
	if len(data) == 0 {
		data = sd.Raw
	}
	return bytes.Contains(data, []byte("pdfaid:part"))
}
