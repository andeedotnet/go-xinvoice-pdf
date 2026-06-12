// Package facturx embeds and extracts the invoice XML of a hybrid ZUGFeRD /
// Factur-X PDF, on top of pdfcpu. It is XML-agnostic: callers pass the already
// serialized CII bytes and receive the embedded bytes back; the go-xinvoice
// dependency lives only in the parent package.
package facturx

import "errors"

const (
	// AttachmentName is the embedded XML file name mandated by Factur-X /
	// ZUGFeRD 2.x.
	AttachmentName = "factur-x.xml"

	// DefaultAFRelationship is the Associated-Files relationship for the
	// BASIC / EN 16931 / XRECHNUNG profiles (the MINIMUM / BASIC WL profiles
	// would use "Data").
	DefaultAFRelationship = "Alternative"
)

// ErrNoEmbeddedXML is returned when a PDF carries no embedded invoice XML.
var ErrNoEmbeddedXML = errors.New("xinvoice-pdf: no embedded invoice XML found")

// legacyNames lists the embedded-file names used by older ZUGFeRD versions, in
// the order extraction prefers them after the current AttachmentName.
var legacyNames = []string{"zugferd-invoice.xml", "xrechnung.xml"}
