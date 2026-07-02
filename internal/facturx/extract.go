package facturx

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// maxDecodedAttachmentBytes caps the *decoded* size of any embedded file
// considered during extraction. Attachments are typically Flate-compressed
// (ratios up to ~1000:1), so without the cap a small crafted PDF could expand
// into gigabytes of memory (decompression bomb). Real invoice XMLs run a few
// MB at most, so 64 MiB is generous. A var so tests can lower it.
var maxDecodedAttachmentBytes int64 = 64 << 20

// embeddedFile is a not-yet-decoded embedded file: the filespec file name and
// the still-encoded stream dict.
type embeddedFile struct {
	name string
	sd   *types.StreamDict
}

// Extract returns the embedded invoice XML and its file name from a hybrid PDF.
// It prefers the Factur-X name (factur-x.xml), then the legacy ZUGFeRD names,
// then any attachment whose content looks like an invoice XML. It returns
// ErrNoEmbeddedXML when nothing matches.
//
// Extraction walks the EmbeddedFiles name tree itself instead of going through
// pdfcpu's attachment API: the latter eagerly decodes every attachment, which
// would defeat the decode cap. Here only the streams actually inspected are
// decoded, one at a time, each limited to maxDecodedAttachmentBytes.
func Extract(pdf []byte) (xml []byte, name string, err error) {
	conf := newConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	// ReadValidateAndOptimize binds the name trees (EmbeddedFiles), which plain
	// ReadContext does not — required before the tree walk below can find them.
	ctx, err := api.ReadValidateAndOptimize(bytes.NewReader(pdf), conf)
	if err != nil {
		return nil, "", fmt.Errorf("xinvoice-pdf: read PDF: %w", err)
	}

	files, err := embeddedFiles(ctx)
	if err != nil || len(files) == 0 {
		return nil, "", ErrNoEmbeddedXML
	}

	// 1. Exact name preference: factur-x.xml, then the legacy names.
	for _, want := range append([]string{AttachmentName}, legacyNames...) {
		for _, f := range files {
			if strings.EqualFold(f.name, want) {
				data, err := decodeEmbeddedFile(f.sd)
				if err != nil {
					return nil, "", fmt.Errorf("xinvoice-pdf: read attachment %q: %w", f.name, err)
				}
				return data, f.name, nil
			}
		}
	}

	// 2. Fall back to any attachment whose content is an invoice XML, decoding
	// one attachment at a time and skipping oversized or broken streams.
	for _, f := range files {
		data, err := decodeEmbeddedFile(f.sd)
		if err != nil {
			continue
		}
		if looksLikeInvoiceXML(data) {
			return data, f.name, nil
		}
	}

	return nil, "", ErrNoEmbeddedXML
}

// embeddedFiles lists the PDF's embedded files without decoding any stream.
// Filespecs that are malformed or carry no embedded stream (references to
// external files) are skipped.
func embeddedFiles(ctx *model.Context) ([]embeddedFile, error) {
	xrt := ctx.XRefTable
	if !xrt.Valid {
		if err := xrt.LocateNameTree("EmbeddedFiles", false); err != nil {
			return nil, err
		}
	}
	tree := xrt.Names["EmbeddedFiles"]
	if tree == nil {
		return nil, nil
	}

	var files []embeddedFile
	collect := func(xrt *model.XRefTable, id string, o *types.Object) error {
		d, err := xrt.DereferenceDict(*o)
		if err != nil || d == nil {
			return nil // malformed filespec: skip
		}
		name := fileSpecName(xrt, d)
		if name == "" {
			name = id
		}
		sd := fileSpecStream(xrt, d)
		if sd == nil {
			return nil // no embedded stream: skip
		}
		files = append(files, embeddedFile{name: name, sd: sd})
		return nil
	}
	if err := tree.Process(xrt, collect); err != nil {
		return nil, err
	}
	return files, nil
}

// fileSpecName returns the filespec's file name (/UF preferred over /F), or ""
// when absent or malformed.
func fileSpecName(xrt *model.XRefTable, d types.Dict) string {
	for _, key := range []string{"UF", "F"} {
		if o, found := d.Find(key); found {
			if s, err := xrt.DereferenceStringOrHexLiteral(o, model.V10, nil); err == nil && s != "" {
				return s
			}
		}
	}
	return ""
}

// fileSpecStream returns the embedded-file stream dict (/EF → /F) of a
// filespec, or nil when absent or malformed.
func fileSpecStream(xrt *model.XRefTable, d types.Dict) *types.StreamDict {
	o, found := d.Find("EF")
	if !found || o == nil {
		return nil
	}
	ef, err := xrt.DereferenceDict(o)
	if err != nil || ef == nil {
		return nil
	}
	o, found = ef.Find("F")
	if !found || o == nil {
		return nil
	}
	sd, _, err := xrt.DereferenceStreamDict(o)
	if err != nil {
		return nil
	}
	return sd
}

// decodeEmbeddedFile returns the decoded content of an embedded-file stream.
// Decoding fails with filter.ErrDecodeLimitExceeded once the decoded output
// exceeds maxDecodedAttachmentBytes.
func decodeEmbeddedFile(sd *types.StreamDict) ([]byte, error) {
	if sd.FilterPipeline == nil {
		return sd.Raw, nil
	}
	if err := sd.DecodeWithLimit(maxDecodedAttachmentBytes); err != nil {
		return nil, err
	}
	return sd.Content, nil
}
