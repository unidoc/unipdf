/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */
package creator

import (
	"fmt"
	"testing"

	"github.com/unidoc/unipdf/v3/model"
)

func TestListSimple(t *testing.T) {
	red := ColorRGBFrom8bit(255, 0, 0)

	c := New()
	c.NewPage()

	list := c.NewList()

	// Add some paragraphs to the list.
	items := []string{
		"Apples",
		"Oranges",
		"Apricots",
		"Cherries",
		"Cranberries",
		"Grapes",
		"Lemons",
	}

	for _, item := range items {
		sp := c.NewStyledParagraph()
		sp.Append(item).Style.Color = red

		list.Add(sp)
	}

	// Shortcut used to add paragraphs to the list.
	list.AddTextItem("Prunes")
	list.AddTextItem("Watermelons")
	list.AddTextItem("Pineapples")
	list.AddTextItem("Clementines")

	err := c.Draw(list)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Write output file.
	err = c.WriteToFile(tempFile("list_simple.pdf"))
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}

func TestListAdvanced(t *testing.T) {
	fontHelveticaBold := model.NewStandard14FontMustCompile(model.HelveticaBoldName)

	red := ColorRGBFrom8bit(255, 0, 0)
	blue := ColorRGBFrom8bit(0, 0, 255)

	c := New()
	c.NewPage()

	list := c.NewList()

	// Add list item.
	list.AddTextItem("Fruit:")

	items := []string{
		"Apple",
		"Orange",
		"Apricot",
		"Cherry",
		"Cranberry",
		"Grape",
		"Lemon",
	}

	// Add a new list to the main list.
	fruitList := c.NewList()
	fruitList.Marker().Text = "- "
	for _, item := range items {
		sp := c.NewStyledParagraph()
		sp.Append(item).Style.Color = red

		fruitList.Add(sp)
	}

	list.Add(fruitList)

	// Add list item.
	list.AddTextItem("Vegetables:")

	items = []string{
		"Chilly",
		"Tomato",
		"Potato",
		"Cucumber",
		"Ginger",
		"Garlic",
		"Onion",
	}

	// Add another list to the main list.
	vegList := c.NewList()
	vegList.Marker().Style.Font = fontHelveticaBold

	for i, item := range items {
		sp := c.NewStyledParagraph()
		sp.Append(item).Style.Color = blue

		marker, _ := vegList.Add(sp)
		marker.Text = fmt.Sprintf("%d. ", i+1)
	}

	// Hide the marker for the newly added list item.
	marker, _ := list.Add(vegList)
	marker.Text = ""

	// Add a long item to the list.
	list.AddTextItem("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.")

	// Add nested lists
	nestedList := c.NewList()

	currentList := nestedList
	for i := 0; i < 5; i++ {
		l := c.NewList()
		l.Marker().Text = "- "
		l.SetIndent(10)

		l.AddTextItem(fmt.Sprintf("Nesting level %d", i+1))
		l.AddTextItem(fmt.Sprintf("Nesting level %d", i+1))
		l.AddTextItem(fmt.Sprintf("Nesting level %d", i+1))

		currentList.Add(l)
		currentList = l
	}

	list.Add(nestedList)

	err := c.Draw(list)
	if err != nil {
		t.Fatalf("Error drawing: %v", err)
	}

	// Write output file.
	err = c.WriteToFile(tempFile("list_advanced.pdf"))
	if err != nil {
		t.Fatalf("Fail: %v\n", err)
	}
}
