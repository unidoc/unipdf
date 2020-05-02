package creator

import (
	"fmt"
	"testing"

	"github.com/unidoc/unipdf/v3/model"
)

func TestInvoiceSimple(t *testing.T) {
	c := New()
	c.NewPage()

	logo, err := c.NewImageFromFile(testImageFile1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
	}

	invoice := c.NewInvoice()

	// Set invoice logo.
	invoice.SetLogo(logo)

	// Set invoice information.
	invoice.SetNumber("0001")
	invoice.SetDate("28/07/2016")
	invoice.SetDueDate("28/07/2016")
	invoice.AddInfo("Payment terms", "Due on receipt")
	invoice.AddInfo("Paid", "No")

	// Set invoice addresses.
	invoice.SetSellerAddress(&InvoiceAddress{
		Heading: "John Doe",
		Street:  "8 Elm Street",
		City:    "Cambridge",
		Zip:     "CB14DH",
		Country: "United Kingdom",
		Phone:   "xxx-xxx-xxxx",
		Email:   "johndoe@email.com",
	})

	invoice.SetBuyerAddress(&InvoiceAddress{
		Heading: "Bill to",
		Name:    "Jane Doe",
		Street:  "9 Elm Street",
		City:    "London",
		Zip:     "LB15FH",
		Country: "United Kingdom",
		Phone:   "xxx-xxx-xxxx",
		Email:   "janedoe@email.com",
	})

	// Add invoice line items.
	for i := 0; i < 75; i++ {
		invoice.AddLine(
			fmt.Sprintf("Test product #%d", i+1),
			"1",
			"$10",
			"$10",
		)
	}

	// Set invoice totals.
	invoice.SetSubtotal("$100.00")
	invoice.AddTotalLine("Tax (10%)", "$10.00")
	invoice.AddTotalLine("Shipping", "$5.00")
	invoice.SetTotal("$115.00")

	// Set invoice content sections.
	invoice.SetNotes("Notes", "Thank you for your business.")
	invoice.SetTerms("Terms and conditions", "Full refund for 60 days after purchase.")

	if err = c.Draw(invoice); err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Write output file.
	err = c.WriteToFile(tempFile("invoice_simple.pdf"))
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

func TestInvoiceAdvanced(t *testing.T) {
	fontHelvetica := model.NewStandard14FontMustCompile(model.HelveticaName)

	c := New()
	c.NewPage()

	logo, err := c.NewImageFromFile(testImageFile1)
	if err != nil {
		t.Errorf("Fail: %v\n", err)
	}

	white := ColorRGBFrom8bit(255, 255, 255)
	lightBlue := ColorRGBFrom8bit(217, 240, 250)
	blue := ColorRGBFrom8bit(2, 136, 209)

	invoice := c.NewInvoice()

	// Set invoice title.
	invoice.SetTitle("Unidoc Invoice")

	// Customize invoice title style.
	titleStyle := invoice.TitleStyle()
	titleStyle.Color = blue
	titleStyle.Font = fontHelvetica
	titleStyle.FontSize = 30

	invoice.SetTitleStyle(titleStyle)

	// Set invoice logo.
	invoice.SetLogo(logo)

	// Set invoice information.
	invoice.SetNumber("0001")
	invoice.SetDate("28/07/2016")
	invoice.SetDueDate("28/07/2016")
	invoice.AddInfo("Payment terms", "Due on receipt")
	invoice.AddInfo("Paid", "No")

	// Customize invoice information styles.
	for _, info := range invoice.InfoLines() {
		descCell, contentCell := info[0], info[1]
		descCell.BackgroundColor = lightBlue
		contentCell.TextStyle.Font = fontHelvetica
	}

	// Set invoice addresses.
	invoice.SetSellerAddress(&InvoiceAddress{
		Heading: "JOHN DOE",
		Street:  "8 Elm Street",
		City:    "Cambridge",
		Zip:     "CB14DH",
		Country: "United Kingdom",
		Phone:   "xxx-xxx-xxxx",
		Email:   "johndoe@email.com",
	})

	invoice.SetBuyerAddress(&InvoiceAddress{
		Heading:   "JANE DOE",
		Name:      "Jane Doe and Associates",
		Street:    "Suite #134569",
		Street2:   "1960 W CHELSEA AVE STE 2006R",
		City:      "ALLENTOWN",
		State:     "PA",
		Zip:       "18104",
		Country:   "United States",
		Phone:     "xxx-xxx-xxxx",
		Email:     "janedoe@email.com",
		Separator: " ",
	})

	// Customize address styles.
	addressStyle := invoice.AddressStyle()
	addressStyle.Font = fontHelvetica
	addressStyle.FontSize = 9

	addressHeadingStyle := invoice.AddressHeadingStyle()
	addressHeadingStyle.Color = blue
	addressHeadingStyle.Font = fontHelvetica
	addressHeadingStyle.FontSize = 16

	invoice.SetAddressStyle(addressStyle)
	invoice.SetAddressHeadingStyle(addressHeadingStyle)

	// Insert new column.
	col := invoice.InsertColumn(2, "Discount")
	col.Alignment = CellHorizontalAlignmentRight

	// Customize column styles.
	for _, column := range invoice.Columns() {
		column.BackgroundColor = lightBlue
		column.BorderColor = lightBlue
		column.TextStyle.FontSize = 9
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
			cell.TextStyle.FontSize = 9
		}
	}

	// Customize total line styles.
	titleCell, contentCell := invoice.Total()
	titleCell.BackgroundColor = lightBlue
	titleCell.BorderColor = lightBlue
	contentCell.BackgroundColor = lightBlue
	contentCell.BorderColor = lightBlue

	invoice.SetSubtotal("$100.00")
	invoice.AddTotalLine("Tax (10%)", "$10.00")
	invoice.AddTotalLine("Shipping", "$5.00")
	invoice.SetTotal("$85.00")

	// Set invoice content sections.
	invoice.SetNotes("NOTES", "Thank you for your business.")
	invoice.SetTerms("I. TERMS OF PAYMENT", "Net 30 days on all invoices. In addition, Buyer shall pay all sales, use, customs, excise or other taxes presently or hereafter payable in regards to this transaction, and Buyer shall reimburse Seller for any such taxes or charges paid by the Seller.\nSeller shall have the continuing right to approve Buyer’s credit. Seller may at any time demand advance payment, additional security or guarantee of prompt payment. If Buyer refuses to give the payment, security or guarantee demanded, Seller may terminate the Agreement, refuse to deliver any undelivered goods and Buyer shall immediately become liable to Seller for the unpaid price of all goods delivered & for damages as provided in Paragraph V below. Buyer agrees to pay Seller cost of collection of overdue invoices, including reasonable attorney’s fees incurred by Seller in collecting said sums. F.O.B. point shall be point of SHIP TO on face hereof.")
	invoice.AddSection("II. DELIVERY, TOLERANCES, WEIGHT", "Upon due tender of goods for delivery at the F.O.B. point all risk of loss or damage and other incident of ownership pass to Buyer, but Seller retains a security interest in the goods until purchase price is paid. All deliveries are subject to weight at shipping point which shall govern.")
	invoice.AddSection("III. WARRANTIES", "Seller warrants that goods sold hereunder are merchantable UNLESS manufactured in conformance with Buyer’s particular specification, and that Seller conveys good title thereto. IN NO EVENT WILL SELLER BE LIABLE FOR CONSEQUENTIAL DAMAGES EVEN IF CUSTOMER HAS NOT BEEN ADVISED OF THE POSSIBILITY OF SUCH DAMAGES. EXCEPT FOR THE EXPRESS WARRANTY STATED IN THIS PARAGRAPH IV, SELLER GRANTS NO WARRANTIES, EITHER EXPRESS OR IMPLIED HEREIN, INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE, AND THIS STATED EXPRESS WARRANTY IS IN LIEU OF ALL LIABILITIES OR OBLIGATIONS OF SELLER FOR DAMAGES INCLUDING BUT NOT LIMITED TO, CONSEQUENTIAL DAMAGES OCCURRING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF ANY GOODS SOLD HEREUNDER. Seller specifically does not warrant the accuracy of sufficiency of any advice or recommendations given to Buyer in connection with the sale of goods hereunder.")
	invoice.AddSection("IV. FORCE MAJEURE", "Seller shall not be liable for any damages resulting from: any delay or failure of performance arising from any cause not reasonably within Seller’s control; accidents to, breakdowns or mechanical failure of machinery or equipment, however caused; strikes or other labor troubles, shortage of labor, transportation, raw materials, energy sources, or failure of ususal means of supply; fire; flood; war, declared or undeclared; insurrection; riots; acts of God or the public enemy; or priorities, allocations or limitations or other acts required or requested by Federal, State or local governments or any of their sub-divisions, bureaus or agencies. Seller may, at its option, cancel this Agreement or delay performance hereunder for any period reasonably necessary due to any of the foregoing, during which time this Agreement shall remain in full force and effect. Seller shall have the further right to then allocate its available goods between its own uses and its customers in such manner as Seller may consider equitable.")
	invoice.AddSection("V. PATENT INDEMNITY", "Seller shall defend and hold Buyer harmless for any action against Seller based in a claim that Buyer’s sale or use of goods normally offered for sale by Seller, supplied by Seller hereunder, and while in the form, state or conditions supplies constitutes infringement of any United States letters patent; provided Seller shall receive prompt written notice of the claim or action, and Buyer shall give Seller authority, information and assistance at Seller’s expense. Buyer shall defend and hold Seller harmless for any action against Seller or its suppliers based in a claim that the manufacture or sale of goods hereunder constitutes infringement of any United States letters patent, if such goods were manufactured pursuant to Buyer’s designs, specifications and /or formulae, and were not normally offered for sale by Seller; provided Buyer shall receive prompt written notice of the claim or action and Seller shall give Buyer authority, information and assistance at Buyer’s expense. Buyer and Seller agree that the foregoing constitutes the parties’ entire liability for claims or actions based on patent infringement.")

	// Customize note styles.
	noteStyle := invoice.NoteStyle()
	noteStyle.Font = fontHelvetica
	noteStyle.FontSize = 12

	noteHeadingStyle := invoice.NoteHeadingStyle()
	noteHeadingStyle.Color = blue
	noteHeadingStyle.Font = fontHelvetica
	noteHeadingStyle.FontSize = 14

	invoice.SetNoteStyle(noteStyle)
	invoice.SetNoteHeadingStyle(noteHeadingStyle)

	if err = c.Draw(invoice); err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Write output file.
	err = c.WriteToFile(tempFile("invoice_advanced.pdf"))
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}
