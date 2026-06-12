package facturx

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/andeedotnet/go-xinvoice-pdf/internal/xmp"
)

func testdata(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return b
}

// TestEmbedExtractRoundTrip embeds the CII fixture into the minimal PDF, proves
// pdfcpu re-reads/validates the result, and extracts the same XML back.
func TestEmbedExtractRoundTrip(t *testing.T) {
	pdf := testdata(t, "minimal.pdf")
	cii := testdata(t, "factur-x.xml")

	out, warnings, err := Embed(pdf, cii, Options{ConformanceLevel: xmp.ConformanceXRechnung})
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	// minimal.pdf is not PDF/A, so a warning is expected (and must not be fatal).
	if len(warnings) == 0 {
		t.Errorf("expected a non-PDF/A warning for the minimal fixture")
	}

	// pdfcpu must accept what we wrote.
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed
	if _, err := api.ReadAndValidate(bytes.NewReader(out), conf); err != nil {
		t.Fatalf("output failed pdfcpu read/validate: %v", err)
	}

	gotXML, name, err := Extract(out)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if name != AttachmentName {
		t.Errorf("extracted name = %q, want %q", name, AttachmentName)
	}
	if !bytes.Equal(gotXML, cii) {
		t.Errorf("extracted XML differs from embedded (%d vs %d bytes)", len(gotXML), len(cii))
	}
}

// TestEmbedWritesFacturXStructures checks the PDF/A-3 associated-file wiring and
// the XMP metadata on the embedded output.
func TestEmbedWritesFacturXStructures(t *testing.T) {
	out, _, err := Embed(testdata(t, "minimal.pdf"), testdata(t, "factur-x.xml"),
		Options{ConformanceLevel: xmp.ConformanceEN16931})
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}

	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed
	ctx, err := api.ReadContext(bytes.NewReader(out), conf)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	xrt := ctx.XRefTable
	root, err := xrt.Catalog()
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	// Catalog /AF -> filespec with /AFRelationship Alternative.
	af := root.ArrayEntry("AF")
	if len(af) == 0 {
		t.Fatal("catalog /AF array missing or empty")
	}
	fs, err := xrt.DereferenceDict(af[0])
	if err != nil {
		t.Fatalf("deref filespec: %v", err)
	}
	if rel := fs.NameEntry("AFRelationship"); rel == nil || *rel != DefaultAFRelationship {
		t.Errorf("filespec /AFRelationship = %v, want %q", rel, DefaultAFRelationship)
	}

	// Embedded-file stream /Subtype text/xml.
	ef, err := xrt.DereferenceDict(fs["EF"])
	if err != nil {
		t.Fatalf("deref EF: %v", err)
	}
	sd, _, err := xrt.DereferenceStreamDict(ef["F"])
	if err != nil || sd == nil {
		t.Fatalf("deref embedded stream: %v", err)
	}
	if st := sd.NameEntry("Subtype"); st == nil || *st != "text/xml" {
		t.Errorf("embedded stream /Subtype = %v, want text/xml", st)
	}

	// Catalog /Metadata -> uncompressed XMP carrying the Factur-X fields.
	mo, found := root.Find("Metadata")
	if !found {
		t.Fatal("catalog /Metadata missing")
	}
	msd, _, err := xrt.DereferenceStreamDict(mo)
	if err != nil || msd == nil {
		t.Fatalf("deref metadata stream: %v", err)
	}
	if len(msd.FilterPipeline) != 0 {
		t.Errorf("XMP metadata stream must be uncompressed, got filters %v", msd.FilterPipeline)
	}
	meta := metadataBytes(t, msd)
	for _, want := range []string{
		xmp.FacturXNamespace,
		"pdfaid:part>3<",
		"<fx:ConformanceLevel>EN 16931</fx:ConformanceLevel>",
	} {
		if !bytes.Contains(meta, []byte(want)) {
			t.Errorf("XMP metadata missing %q", want)
		}
	}
}

func metadataBytes(t *testing.T, sd *types.StreamDict) []byte {
	t.Helper()
	if len(sd.Content) > 0 {
		return sd.Content
	}
	if len(sd.FilterPipeline) == 0 {
		return sd.Raw
	}
	if err := sd.Decode(); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}
	return sd.Content
}

// TestExtractLegacyFilename verifies the extraction fallback to the legacy
// ZUGFeRD attachment name when factur-x.xml is absent.
func TestExtractLegacyFilename(t *testing.T) {
	pdf := testdata(t, "minimal.pdf")
	cii := testdata(t, "factur-x.xml")

	// Embed under a legacy name by attaching directly with that file name.
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed
	conf.WriteObjectStream = false
	conf.WriteXRefStream = false
	ctx, err := api.ReadContext(bytes.NewReader(pdf), conf)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	a := model.Attachment{Reader: bytes.NewReader(cii), ID: "zugferd-invoice.xml", FileName: "zugferd-invoice.xml"}
	if err := ctx.AddAttachment(a, false); err != nil {
		t.Fatalf("AddAttachment: %v", err)
	}
	var buf bytes.Buffer
	if err := api.WriteContext(ctx, &buf); err != nil {
		t.Fatalf("write: %v", err)
	}

	gotXML, name, err := Extract(buf.Bytes())
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if name != "zugferd-invoice.xml" {
		t.Errorf("extracted name = %q, want zugferd-invoice.xml", name)
	}
	if !bytes.Equal(gotXML, cii) {
		t.Errorf("extracted XML differs from embedded")
	}
}

// TestExtractNoAttachment reports ErrNoEmbeddedXML for a plain PDF.
func TestExtractNoAttachment(t *testing.T) {
	if _, _, err := Extract(testdata(t, "minimal.pdf")); err != ErrNoEmbeddedXML {
		t.Errorf("Extract on plain PDF: err = %v, want ErrNoEmbeddedXML", err)
	}
}
