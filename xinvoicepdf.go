package xinvoicepdf

import (
	"fmt"
	"strings"

	xinvoice "github.com/andeedotnet/go-xinvoice"

	"github.com/andeedotnet/go-xinvoice-pdf/internal/facturx"
	"github.com/andeedotnet/go-xinvoice-pdf/internal/xmp"
)

// Profile selects the Factur-X / ZUGFeRD conformance level written to the XMP
// metadata. The embedded syntax is always CII regardless of profile.
type Profile int

const (
	// ProfileXRechnung writes the "XRECHNUNG" conformance level (default).
	ProfileXRechnung Profile = iota
	// ProfileEN16931 writes the "EN 16931" conformance level.
	ProfileEN16931
)

func (p Profile) conformanceLevel() string {
	if p == ProfileEN16931 {
		return xmp.ConformanceEN16931
	}
	return xmp.ConformanceXRechnung
}

// String returns the CLI profile name ("xrechnung" / "en16931").
func (p Profile) String() string {
	if p == ProfileEN16931 {
		return "en16931"
	}
	return "xrechnung"
}

// ParseProfile maps a CLI profile name to a Profile.
func ParseProfile(s string) (Profile, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "xrechnung", "xr":
		return ProfileXRechnung, nil
	case "en16931", "en 16931", "comfort", "en":
		return ProfileEN16931, nil
	default:
		return 0, fmt.Errorf("xinvoice-pdf: unknown profile %q (want xrechnung|en16931)", s)
	}
}

// Options controls embedding. The zero value embeds with the XRECHNUNG profile
// and the default "Alternative" associated-file relationship.
type Options struct {
	Profile        Profile
	AFRelationship string // filespec /AFRelationship; default "Alternative"
	Description    string // filespec /Desc; default "Factur-X invoice"
}

// ErrNoEmbeddedXML is returned by ExtractXML / ExtractInvoice when a PDF carries
// no embedded invoice XML.
var ErrNoEmbeddedXML = facturx.ErrNoEmbeddedXML

// Embed inserts the CII invoice XML into pdf and returns a hybrid Factur-X PDF.
// ciiXML must already be UN/CEFACT CII (use EmbedInvoice to serialize a model).
// warnings is non-empty when pdf is not already PDF/A — the output is still a
// valid Factur-X hybrid, but strict PDF/A-3 validity requires a PDF/A input.
func Embed(pdf, ciiXML []byte, opt *Options) (out []byte, warnings []string, err error) {
	if opt == nil {
		opt = &Options{}
	}
	return facturx.Embed(pdf, ciiXML, facturx.Options{
		ConformanceLevel: opt.Profile.conformanceLevel(),
		AFRelationship:   opt.AFRelationship,
		Description:      opt.Description,
	})
}

// EmbedInvoice serializes inv to CII and embeds it via Embed.
func EmbedInvoice(pdf []byte, inv *xinvoice.Invoice, opt *Options) (out []byte, warnings []string, err error) {
	cii, err := inv.ToXML(xinvoice.CII)
	if err != nil {
		return nil, nil, fmt.Errorf("xinvoice-pdf: serialize CII: %w", err)
	}
	return Embed(pdf, cii, opt)
}

// ExtractXML returns the embedded invoice XML and its file name from a hybrid PDF.
func ExtractXML(pdf []byte) (xml []byte, name string, err error) {
	return facturx.Extract(pdf)
}

// ExtractInvoice extracts the embedded XML and parses it (auto-detecting UBL/CII)
// into a go-xinvoice Invoice.
func ExtractInvoice(pdf []byte) (*xinvoice.Invoice, error) {
	xml, _, err := facturx.Extract(pdf)
	if err != nil {
		return nil, err
	}
	return xinvoice.ParseXML(xml)
}
