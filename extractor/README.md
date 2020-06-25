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

`text_page.go` **makeTextPage()** is the top level text extraction function. It returns an ordered
list of `textPara`s which are described below.

* A page's `textMark`s are obtained from its content stream. They are in the order they occur in the content stream.
* The `textMark`s are grouped into word fragments called`textWord`s by scanning through the textMarks
 and splitting on space characters and the gaps between marks.
* The `textWords`s are grouped into rectangular regions  based on their bounding boxes' proximities
  to other `textWords`. These rectangular regions are called `textParas`s. (In the current implementation
  there is an intermediate step where the `textWords` are divided into containers called `wordBags`.)
* The `textWord`s in each `textPara` are arranged into `textLine`s (`textWord`s of similar depth).
* Within each `textLine`, `textWord`s are sorted in reading order and each one that starts a whole
word is marked by setting its `newWord` flag to true. (See `textLine.text()`.)
* All the `textPara`s on a page are checked to see if they are arranged as cells within a table and,
if they are, they are combined into `textTable`s and a `textPara` containing the `textTable` replaces
the `textPara`s containing the cells.
* The `textPara`s, some of which may be tables, are sorted into reading order (the order in which they
are read, not in the *reading* direction).


The entire order of extracted text from a page is expressed in `paraList.writeText()`.

* This function iterates through the `textPara`s, which are sorted in reading order.
* For each `textPara` with a table, it iterates through the table cell `textPara`s. (See
 `textPara.writeCellText()`.)
* For each (top level or table cell) `textPara`, it iterates through the `textLine`s.
* For each `textLine`, it iterates through the `textWord`s inserting a space before each one that has
 the `newWord` flag set.


### `textWord` creation

* `makeTextWords()` combines `textMark`s into `textWord`s, word fragments.
* `textWord`s are the atoms of the text extraction code.

### `textPara` creation

* `dividePage()` combines `textWord`s that are close to each other into groups in rectangular
 regions called `wordBags`.
* `wordBag.arrangeText()` arranges the `textWord`s in the rectangular regions into `textLine`s,
  groups textWords of about the same depth sorted left to right.
* `textLine.markWordBoundaries()` marks the `textWord`s in each `textLine` that start whole words.

TODO
-----

* Handle diagonal text.
* Get R to L text extraction working.
* Get top to bottom text extraction working.
