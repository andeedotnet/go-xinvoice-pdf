package xinvoicepdf_test

import (
	"fmt"
	"log"

	xinvoicepdf "github.com/andeedotnet/go-xinvoice-pdf"
)

// ExampleEmbed embeds a CII invoice XML into an existing PDF and reads it back.
func ExampleEmbed() {
	var pdf, ciiXML []byte // a PDF you provide and the CII invoice XML

	hybrid, warnings, err := xinvoicepdf.Embed(pdf, ciiXML, nil) // nil => XRECHNUNG profile
	if err != nil {
		log.Fatal(err)
	}
	for _, w := range warnings {
		log.Println("warning:", w)
	}

	xml, name, err := xinvoicepdf.ExtractXML(hybrid)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("embedded %s (%d bytes)\n", name, len(xml))
}
