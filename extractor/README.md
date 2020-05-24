TEXT EXTRACTION CODE
====================
The code is currently split accross the text_*.go files to make it easier to navigate. Once you
understand the code you may wish to recombine this in the orginal text.go
\

BASIC IDEAS
-----------
There are two directions

- *reading*
- *depth*

In English text,
- the *reading* direction is left to right, increasing X in the PDF coordinate system.
- the *depth* directon is top to bottom, decreasing Y in the PDF coordinate system.

We define *depth* as distance from the bottom of a word's bounding box from the top of the page.
depth := pageSize.Ury - r.Lly

* Pages are divided into rectangular regions called `textPara`s.
* The `textPara`s in a page are sorted in reading ouder (the order they are read, not the
*reading* direction above).
* Each `textPara` contains `textLine`s, lines with the `textPara`'s bounding box.
* Each `textLine` has a text reprentation.

Page text is extracted by iterating over `textPara`s and within each `textPara` iterating over its
`textLine`s.


WHERE TO START
--------------

`text_page.go` *makeTextPage* is the top level function that builds the `textPara`s.

* A page's `textMark`s are obtained from its contentstream.
* The `textMark`s are divided into `textWord`s.
* The `textWord`s are grouped into depth bins with each the contents of each bin sorted by reading direction.
* The page area is into rectangular regions for each paragraph.
* The words in of each rectangular region are aranged inot`textLine`s. Each rectangular region and
its constituent lines is a `textPara`.
* The `textPara`s are sorted into reading order.


