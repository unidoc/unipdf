package creator

import (
	"fmt"
	"testing"
)

func TestInvoiceSimple(t *testing.T) {
	c := New()
	c.NewPage()

	logo, err := c.NewImageFromFile(testImageFile1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
	}

	invoice := c.NewInvoice()

	invoice.SetLogo(logo)
	invoice.SetNumber("0001")
	invoice.SetDate("28/07/2016")
	invoice.SetDueDate("28/07/2016")
	invoice.AddInvoiceInfo("Payment terms", "Due on receipt")
	invoice.AddInvoiceInfo("Paid", "No")

	invoice.SetSellerAddress(&InvoiceAddress{
		Name:    "John Doe",
		Street:  "8 Elm Street",
		City:    "Cambridge",
		Zip:     "CB14DH",
		Country: "United Kingdom",
		Phone:   "xxx-xxx-xxxx",
		Email:   "johndoe@email.com",
	})

	invoice.SetBuyerAddress(&InvoiceAddress{
		Name:    "Jane Doe",
		Street:  "9 Elm Street",
		City:    "London",
		Zip:     "LB15FH",
		Country: "United Kingdom",
		Phone:   "xxx-xxx-xxxx",
		Email:   "janedoe@email.com",
	})

	for i := 0; i < 10; i++ {
		invoice.AddLine(
			fmt.Sprintf("Test product #%d", i+1),
			"1",
			"$10",
			"$10",
		)
	}

	invoice.SetSubtotal("$100.00")
	invoice.AddTotalLine("Tax (10%)", "$10.00")
	invoice.AddTotalLine("Shipping", "$5.00")
	invoice.SetTotal("$115.00")
	invoice.SetNotes("Notes", "Thank you for your business")
	invoice.SetTerms("Terms and conditions", "Full refund for 60 days after purchase.")

	if err = c.Draw(invoice); err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Write output file.
	err = c.WriteToFile("/tmp/invoice_simple.pdf")
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}
