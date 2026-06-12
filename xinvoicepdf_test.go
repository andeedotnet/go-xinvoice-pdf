package xinvoicepdf_test

import (
	"bytes"
	"os"
	"testing"

	xinvoice "github.com/andeedotnet/go-xinvoice"
	xinvoicepdf "github.com/andeedotnet/go-xinvoice-pdf"
)

func read(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return b
}

func TestEmbedThenExtractXML(t *testing.T) {
	pdf := read(t, "minimal.pdf")
	cii := read(t, "factur-x.xml")

	out, _, err := xinvoicepdf.Embed(pdf, cii, nil) // nil opts => XRECHNUNG
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	got, name, err := xinvoicepdf.ExtractXML(out)
	if err != nil {
		t.Fatalf("ExtractXML: %v", err)
	}
	if name != "factur-x.xml" {
		t.Errorf("name = %q, want factur-x.xml", name)
	}
	if !bytes.Equal(got, cii) {
		t.Errorf("round-tripped XML differs from input")
	}
}

func TestEmbedInvoiceExtractInvoice(t *testing.T) {
	pdf := read(t, "minimal.pdf")

	inv, err := xinvoice.ParseXML(read(t, "factur-x.xml"))
	if err != nil {
		t.Fatalf("ParseXML fixture: %v", err)
	}

	out, _, err := xinvoicepdf.EmbedInvoice(pdf, inv, &xinvoicepdf.Options{Profile: xinvoicepdf.ProfileEN16931})
	if err != nil {
		t.Fatalf("EmbedInvoice: %v", err)
	}

	got, err := xinvoicepdf.ExtractInvoice(out)
	if err != nil {
		t.Fatalf("ExtractInvoice: %v", err)
	}
	if got.Number != inv.Number || got.Number == "" {
		t.Errorf("round-tripped invoice number = %q, want %q", got.Number, inv.Number)
	}
}

func TestParseProfile(t *testing.T) {
	cases := map[string]xinvoicepdf.Profile{
		"":          xinvoicepdf.ProfileXRechnung,
		"xrechnung": xinvoicepdf.ProfileXRechnung,
		"EN16931":   xinvoicepdf.ProfileEN16931,
		"en 16931":  xinvoicepdf.ProfileEN16931,
	}
	for in, want := range cases {
		got, err := xinvoicepdf.ParseProfile(in)
		if err != nil {
			t.Errorf("ParseProfile(%q): %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ParseProfile(%q) = %v, want %v", in, got, want)
		}
	}
	if _, err := xinvoicepdf.ParseProfile("bogus"); err == nil {
		t.Errorf("ParseProfile(bogus) should error")
	}
}
