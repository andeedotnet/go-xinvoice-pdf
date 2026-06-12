package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// fixtures live at the module root; tests run with cwd = this package dir.
const (
	pdfFixture = "../../testdata/minimal.pdf"
	xmlFixture = "../../testdata/factur-x.xml"
)

func TestCLIEmbedExtractRoundTrip(t *testing.T) {
	dir := t.TempDir()
	hybrid := filepath.Join(dir, "hybrid.pdf")
	outXML := filepath.Join(dir, "out.xml")

	if code := cmdEmbed([]string{"--pdf", pdfFixture, "--in", xmlFixture, "--out", hybrid, "--profile", "xrechnung"}); code != 0 {
		t.Fatalf("embed exit = %d, want 0", code)
	}
	if code := cmdExtract([]string{"--in", hybrid, "--out", outXML}); code != 0 {
		t.Fatalf("extract exit = %d, want 0", code)
	}

	got, err := os.ReadFile(outXML)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := os.ReadFile(xmlFixture)
	if !bytes.Equal(got, want) {
		t.Errorf("CLI round-trip XML mismatch (%d vs %d bytes)", len(got), len(want))
	}
}

func TestCLIEmbedUsageErrors(t *testing.T) {
	if code := cmdEmbed([]string{"--in", xmlFixture}); code != 2 {
		t.Errorf("embed without --pdf: exit = %d, want 2", code)
	}
	if code := cmdEmbed([]string{"--pdf", pdfFixture, "--in", xmlFixture, "--profile", "bogus"}); code != 2 {
		t.Errorf("embed with bad --profile: exit = %d, want 2", code)
	}
}

func TestCLIExtractNoAttachment(t *testing.T) {
	if code := cmdExtract([]string{"--in", pdfFixture, "--out", filepath.Join(t.TempDir(), "x.xml")}); code != 1 {
		t.Errorf("extract on plain PDF: exit = %d, want 1", code)
	}
}
