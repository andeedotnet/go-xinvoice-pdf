package xmp

import (
	"bytes"
	"strings"
	"testing"
)

func TestPacketContainsRequiredFields(t *testing.T) {
	p := Packet(ConformanceXRechnung)

	if !bytes.HasPrefix(p, append([]byte(`<?xpacket begin="`), utf8BOM...)) {
		t.Errorf("packet must start with an xpacket PI carrying the UTF-8 BOM")
	}
	for _, want := range []string{
		FacturXNamespace,
		"<pdfaid:part>3</pdfaid:part>",
		"<pdfaid:conformance>B</pdfaid:conformance>",
		"<fx:DocumentType>INVOICE</fx:DocumentType>",
		"<fx:DocumentFileName>factur-x.xml</fx:DocumentFileName>",
		"<fx:Version>1.0</fx:Version>",
		"<fx:ConformanceLevel>XRECHNUNG</fx:ConformanceLevel>",
		"pdfaExtension:schemas", // the extension-schema block the fx: props require
		`<?xpacket end="w"?>`,
	} {
		if !strings.Contains(string(p), want) {
			t.Errorf("packet missing %q", want)
		}
	}
}

func TestPacketConformanceLevelPerProfile(t *testing.T) {
	if !strings.Contains(string(Packet(ConformanceEN16931)), "<fx:ConformanceLevel>EN 16931</fx:ConformanceLevel>") {
		t.Errorf("EN16931 profile must write the 'EN 16931' conformance level")
	}
}
