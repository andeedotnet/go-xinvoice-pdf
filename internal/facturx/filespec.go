package facturx

import (
	"bytes"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// attachInvoice embeds data as the factur-x.xml associated file: it registers
// the embedded stream and its filespec in the EmbeddedFiles name tree, sets the
// Factur-X /AFRelationship, and links the filespec from the catalog /AF array.
func attachInvoice(xrt *model.XRefTable, data []byte, opt Options) error {
	if err := xrt.LocateNameTree("EmbeddedFiles", true); err != nil {
		return err
	}

	streamRef, err := newXMLEmbeddedStream(xrt, data, opt.ModTime)
	if err != nil {
		return err
	}

	fsDict, err := xrt.NewFileSpecDict(AttachmentName, AttachmentName, opt.Description, *streamRef)
	if err != nil {
		return err
	}
	// Factur-X requires the associated-file relationship on the filespec.
	fsDict.InsertName("AFRelationship", opt.AFRelationship)

	fsRef, err := xrt.IndRefForNewObject(fsDict)
	if err != nil {
		return err
	}

	m := model.NameMap{AttachmentName: []types.Dict{fsDict}}
	if err := xrt.Names["EmbeddedFiles"].Add(xrt, AttachmentName, *fsRef, m, []string{"F", "UF"}); err != nil {
		return err
	}

	return addAssociatedFile(xrt, *fsRef)
}

// newXMLEmbeddedStream builds the embedded-file stream for the invoice XML. It
// mirrors XRefTable.NewEmbeddedStreamDict but additionally sets the Factur-X
// /Subtype text/xml on the stream (written as /text#2Fxml).
func newXMLEmbeddedStream(xrt *model.XRefTable, data []byte, modTime time.Time) (*types.IndirectRef, error) {
	sd, err := xrt.NewStreamDictForBuf(data)
	if err != nil {
		return nil, err
	}
	sd.InsertName("Type", "EmbeddedFile")
	sd.InsertName("Subtype", "text/xml")

	params := types.NewDict()
	params.InsertInt("Size", len(data))
	params.Insert("ModDate", types.StringLiteral(types.DateString(modTime)))
	sd.Insert("Params", params)

	if err := sd.Encode(); err != nil {
		return nil, err
	}
	return xrt.IndRefForNewObject(*sd)
}

// addAssociatedFile appends fsRef to the document catalog /AF array (creating it
// when absent), making the embedded file an associated file per PDF/A-3.
func addAssociatedFile(xrt *model.XRefTable, fsRef types.IndirectRef) error {
	root, err := xrt.Catalog()
	if err != nil {
		return err
	}
	if af := root.ArrayEntry("AF"); af != nil {
		root.Update("AF", append(af, fsRef))
	} else {
		root.Insert("AF", types.Array{fsRef})
	}
	return nil
}

// setMetadata replaces the document catalog /Metadata with an uncompressed XMP
// stream carrying the Factur-X packet. PDF/A wants the metadata stream
// unfiltered, so the stream is written verbatim (no Flate).
func setMetadata(xrt *model.XRefTable, packet []byte) error {
	length := int64(len(packet))
	sd := types.StreamDict{
		Dict:         types.NewDict(),
		Raw:          packet,
		Content:      packet,
		StreamLength: &length,
	}
	sd.InsertName("Type", "Metadata")
	sd.InsertName("Subtype", "XML")
	sd.InsertInt("Length", int(length))

	ref, err := xrt.IndRefForNewObject(sd)
	if err != nil {
		return err
	}
	root, err := xrt.Catalog()
	if err != nil {
		return err
	}
	root.Update("Metadata", *ref)
	return nil
}

// looksLikeInvoiceXML reports whether b is an XML document whose root element is
// a CII CrossIndustryInvoice or a UBL Invoice — used as the last-resort match
// when extracting from a PDF whose attachment is named unconventionally.
func looksLikeInvoiceXML(b []byte) bool {
	t := bytes.TrimSpace(b)
	return bytes.Contains(t, []byte("CrossIndustryInvoice")) || bytes.Contains(t, []byte(":Invoice")) || bytes.HasPrefix(t, []byte("<Invoice"))
}
