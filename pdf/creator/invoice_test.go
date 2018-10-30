package creator

import (
	"fmt"
	"testing"

	"github.com/unidoc/unidoc/pdf/model"
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
	invoice.SetNotes("Notes", "Thank you for your business.")
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

func TestInvoiceAdvanced(t *testing.T) {
	fontHelvetica := model.NewStandard14FontMustCompile(model.Helvetica)

	c := New()
	c.NewPage()

	logo, err := c.NewImageFromFile(testImageFile1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
	}

	white := ColorRGBFrom8bit(255, 255, 255)
	lightBlue := ColorRGBFrom8bit(217, 240, 250)

	invoice := c.NewInvoice()

	invoice.SetLogo(logo)

	// Customize invoice info
	descCell, contentCell := invoice.SetNumber("0001")
	descCell.BackgroundColor = lightBlue
	contentCell.TextStyle.Font = fontHelvetica

	descCell, contentCell = invoice.SetDate("28/07/2016")
	descCell.BackgroundColor = lightBlue
	contentCell.TextStyle.Font = fontHelvetica

	descCell, contentCell = invoice.SetDueDate("28/07/2016")
	descCell.BackgroundColor = lightBlue
	contentCell.TextStyle.Font = fontHelvetica

	descCell, contentCell = invoice.AddInvoiceInfo("Payment terms", "Due on receipt")
	descCell.BackgroundColor = lightBlue
	contentCell.TextStyle.Font = fontHelvetica

	descCell, contentCell = invoice.AddInvoiceInfo("Paid", "No")
	descCell.BackgroundColor = lightBlue
	contentCell.TextStyle.Font = fontHelvetica

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

	// Insert new column
	col := invoice.InsertColumn(2, "Discount")
	col.Alignment = CellHorizontalAlignmentRight

	// Customize column styles.
	for _, column := range invoice.Columns() {
		column.BackgroundColor = lightBlue
		column.BorderColor = lightBlue
	}

	for i := 0; i < 7; i++ {
		cells := invoice.AddLine(
			fmt.Sprintf("Test product #%d", i+1),
			"1",
			"0%",
			"$10",
			"$10",
		)

		for _, cell := range cells {
			cell.BorderColor = white
		}
	}

	// Customize total line style.
	titleCell, contentCell := invoice.Total()
	titleCell.BackgroundColor = lightBlue
	titleCell.BorderColor = lightBlue
	contentCell.BackgroundColor = lightBlue
	contentCell.BorderColor = lightBlue

	invoice.SetSubtotal("$100.00")
	invoice.AddTotalLine("Tax (10%)", "$10.00")
	invoice.AddTotalLine("Shipping", "$5.00")
	invoice.SetTotal("$85.00")

	invoice.SetNotes("Notes", "Thank you for your business.")
	invoice.SetTerms("Terms and conditions", "Full refund for 60 days after purchase.")
	invoice.AddSection("Custom section", "This is a custom section.")

	if err = c.Draw(invoice); err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Write output file.
	err = c.WriteToFile("/tmp/invoice_advanced.pdf")
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}
