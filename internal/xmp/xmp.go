// Package xmp builds the Factur-X / ZUGFeRD XMP metadata packet embedded in a
// hybrid invoice PDF. The packet identifies the file as PDF/A-3b and carries the
// four mandatory Factur-X fields together with the PDF/A extension schema that
// the custom fx: properties require.
package xmp

import (
	"bytes"
	"text/template"
)

// FacturXNamespace is the Factur-X / ZUGFeRD 2.x XMP namespace URI (the trailing
// '#' is part of the URI).
const FacturXNamespace = "urn:factur-x:pdfa:CrossIndustryDocument:invoice:1p0#"

// Conformance levels written into the Factur-X fx:ConformanceLevel field.
const (
	// ConformanceXRechnung is the German XRechnung CIUS reference profile.
	ConformanceXRechnung = "XRECHNUNG"
	// ConformanceEN16931 is the generic EN 16931 (COMFORT) profile.
	ConformanceEN16931 = "EN 16931"
)

// facturXVersion is the Factur-X XML schema version (not the invoice version).
const facturXVersion = "1.0"

// utf8BOM is the byte-order mark that the xpacket "begin" attribute carries.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

var bodyTmpl = template.Must(template.New("xmp").Parse(xmpBody))

// Packet returns the complete XMP metadata packet for the given Factur-X
// conformance level (one of the Conformance* constants).
func Packet(conformanceLevel string) []byte {
	var buf bytes.Buffer
	buf.WriteString(`<?xpacket begin="`)
	buf.Write(utf8BOM)
	buf.WriteString(`" id="W5M0MpCehiHzreSzNTczkc9d"?>` + "\n")
	// bodyTmpl is a static template with a single trusted value, so Execute
	// cannot fail; the error is intentionally ignored.
	_ = bodyTmpl.Execute(&buf, struct {
		Namespace        string
		Version          string
		ConformanceLevel string
	}{FacturXNamespace, facturXVersion, conformanceLevel})
	buf.WriteString("\n" + `<?xpacket end="w"?>`)
	return buf.Bytes()
}

const xmpBody = `<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">

  <rdf:Description rdf:about="" xmlns:pdfaid="http://www.aiim.org/pdfa/ns/id/">
   <pdfaid:part>3</pdfaid:part>
   <pdfaid:conformance>B</pdfaid:conformance>
  </rdf:Description>

  <rdf:Description rdf:about="" xmlns:fx="{{.Namespace}}">
   <fx:DocumentType>INVOICE</fx:DocumentType>
   <fx:DocumentFileName>factur-x.xml</fx:DocumentFileName>
   <fx:Version>{{.Version}}</fx:Version>
   <fx:ConformanceLevel>{{.ConformanceLevel}}</fx:ConformanceLevel>
  </rdf:Description>

  <rdf:Description rdf:about=""
     xmlns:pdfaExtension="http://www.aiim.org/pdfa/ns/extension/"
     xmlns:pdfaSchema="http://www.aiim.org/pdfa/ns/schema#"
     xmlns:pdfaProperty="http://www.aiim.org/pdfa/ns/property#">
   <pdfaExtension:schemas>
    <rdf:Bag>
     <rdf:li rdf:parseType="Resource">
      <pdfaSchema:schema>Factur-X PDFA Extension Schema</pdfaSchema:schema>
      <pdfaSchema:namespaceURI>{{.Namespace}}</pdfaSchema:namespaceURI>
      <pdfaSchema:prefix>fx</pdfaSchema:prefix>
      <pdfaSchema:property>
       <rdf:Seq>
        <rdf:li rdf:parseType="Resource">
         <pdfaProperty:name>DocumentFileName</pdfaProperty:name>
         <pdfaProperty:valueType>Text</pdfaProperty:valueType>
         <pdfaProperty:category>external</pdfaProperty:category>
         <pdfaProperty:description>name of the embedded XML invoice file</pdfaProperty:description>
        </rdf:li>
        <rdf:li rdf:parseType="Resource">
         <pdfaProperty:name>DocumentType</pdfaProperty:name>
         <pdfaProperty:valueType>Text</pdfaProperty:valueType>
         <pdfaProperty:category>external</pdfaProperty:category>
         <pdfaProperty:description>INVOICE</pdfaProperty:description>
        </rdf:li>
        <rdf:li rdf:parseType="Resource">
         <pdfaProperty:name>Version</pdfaProperty:name>
         <pdfaProperty:valueType>Text</pdfaProperty:valueType>
         <pdfaProperty:category>external</pdfaProperty:category>
         <pdfaProperty:description>The actual version of the Factur-X XML schema</pdfaProperty:description>
        </rdf:li>
        <rdf:li rdf:parseType="Resource">
         <pdfaProperty:name>ConformanceLevel</pdfaProperty:name>
         <pdfaProperty:valueType>Text</pdfaProperty:valueType>
         <pdfaProperty:category>external</pdfaProperty:category>
         <pdfaProperty:description>The conformance level of the embedded Factur-X data</pdfaProperty:description>
        </rdf:li>
       </rdf:Seq>
      </pdfaSchema:property>
     </rdf:li>
    </rdf:Bag>
   </pdfaExtension:schemas>
  </rdf:Description>

  <rdf:Description rdf:about="" xmlns:xmp="http://ns.adobe.com/xap/1.0/">
   <xmp:CreatorTool>go-xinvoice-pdf</xmp:CreatorTool>
  </rdf:Description>

 </rdf:RDF>
</x:xmpmeta>`
