package facturx

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// Extract returns the embedded invoice XML and its file name from a hybrid PDF.
// It prefers the Factur-X name (factur-x.xml), then the legacy ZUGFeRD names,
// then any attachment whose content looks like an invoice XML. It returns
// ErrNoEmbeddedXML when nothing matches.
func Extract(pdf []byte) (xml []byte, name string, err error) {
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	// ReadValidateAndOptimize binds the name trees (EmbeddedFiles), which plain
	// ReadContext does not — required before ExtractAttachments can find them.
	ctx, err := api.ReadValidateAndOptimize(bytes.NewReader(pdf), conf)
	if err != nil {
		return nil, "", fmt.Errorf("xinvoice-pdf: read PDF: %w", err)
	}

	attachments, err := ctx.ExtractAttachments(nil)
	if err != nil || len(attachments) == 0 {
		return nil, "", ErrNoEmbeddedXML
	}

	// Read every attachment's bytes once, keyed by (lower-cased) file name.
	type att struct {
		name string
		data []byte
	}
	read := make([]att, 0, len(attachments))
	for _, a := range attachments {
		if a.Reader == nil {
			continue
		}
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, a.Reader); err != nil {
			return nil, "", fmt.Errorf("xinvoice-pdf: read attachment %q: %w", a.FileName, err)
		}
		read = append(read, att{name: a.FileName, data: buf.Bytes()})
	}

	// 1. Exact name preference: factur-x.xml, then the legacy names.
	for _, want := range append([]string{AttachmentName}, legacyNames...) {
		for _, a := range read {
			if strings.EqualFold(a.name, want) {
				return a.data, a.name, nil
			}
		}
	}

	// 2. Fall back to any attachment whose content is an invoice XML.
	for _, a := range read {
		if looksLikeInvoiceXML(a.data) {
			return a.data, a.name, nil
		}
	}

	return nil, "", ErrNoEmbeddedXML
}
