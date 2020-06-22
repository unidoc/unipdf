TEXT EXTRACTION CODE
====================

There are two [directions](https://www.w3.org/International/questions/qa-scripts.en#directions)s\.

- *reading*
- *depth*

In English text,
- the *reading* direction is left to right, increasing X in the PDF coordinate system.
- the *depth* directon is top to bottom, decreasing Y in the PDF coordinate system.


HOW TEXT IS EXTRACTED
---------------------

`text_page.go` **makeTextPage** is the top level function that builds the `textPara`s.

* A page's `textMark`s are obtained from its contentstream. They are in the order they occur in the contentstrem.
* The `textMark`s are grouped into word fragments called`textWord`s by scanning through the textMarks
 and spltting on space characters and the gaps between marks.
* The `textWords`s are grouped into `textParas`s based on their bounding boxes' proximities to other
 textWords.
* The `textWord`s in each `textPara` are arranged into `textLine`s (`textWord`s of similar depth).
* Within each `textLine`, `textWord`s are sorted in reading order each one that starts a whole word is marked.
See `textLine.text()`.
* `textPara.writeCellText()` shows how to extract the paragraph text from this arrangment.
* All the `textPara`s on a page are checked to see if they are arranged as cells within a table and,
if they are, they are combined into `textTable`s and a textPara containing the textTable replaces the
the `textPara`s containing the cells.
* The textParas, some of which may be tables, in sorted into reading order (the order in which they
are reading, not in the reading directions).


### `textWord` creation

* `makeTextWords()` combines `textMark`s into `textWord`s, word fragments
* textWord`s are the atoms of the text extraction code.

### `textPara` creation

* `dividePage()` combines `textWord`s, that are close to each other into groups in rectangular
 regions called `wordBags`.
* wordBag.arrangeText() arranges the textWords in the rectangle into `textLine`s, groups textWords
of about the same depth sorted left to right.
* textLine.markWordBoundaries() marks the textWords in each textLine that start whole words.

TODO
-----

* Remove serial code????
* Remove verbose* logginng?
* Reinstate rotated text handling.
* Reinstate  diacritic composition.
* Reinstate duplicate text removal.
* Reinstate creater_test.go extraction test.

