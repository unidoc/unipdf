TEXT EXTRACTION CODE
====================
The code is currently split accross the `text_*.go` files to make it easier to navigate. Once you
understand the code you may wish to recombine this in the orginal `text.go`.

BASIC IDEAS
-----------
There are two [directions](https://www.w3.org/International/questions/qa-scripts.en#directions)s\.

- *reading*
- *depth*

In English text,
- the *reading* direction is left to right, increasing X in the PDF coordinate system.
- the *depth* directon is top to bottom, decreasing Y in the PDF coordinate system.

We define *depth* as distance from the bottom of a word's bounding box from the top of the page.
depth := pageSize.Ury - r.Lly

* Pages are divided into rectangular regions called `textPara`s.
* The `textPara`s in a page are sorted in reading order (the order they are read in, not the
*reading* direction above).
* Each `textPara` contains `textLine`s, lines with the `textPara`'s bounding box.
* Each `textLine` has extracted for the line in its `text()` function.

Page text is extracted by iterating over `textPara`s and within each `textPara` iterating over its
`textLine`s.


WHERE TO START
--------------

`text_page.go` **makeTextPage** is the top level function that builds the `textPara`s.

* A page's `textMark`s are obtained from its contentstream.
* The `textMark`s are divided into `textWord`s.
* The `textWord`s are grouped into depth bins with the contents of each bin sorted by reading direction.
* The page area is divided into rectangular regions, one for each paragraph.
* The words in of each rectangular region are aranged inot`textLine`s. Each rectangular region and
its constituent lines is a `textPara`.
* The `textPara`s are sorted into reading order.


TODO
====
Remove serial code????
Reinstate rotated text handling.
Reinstate hyphen diacritic composition.
Reinstate duplicate text removal
Get these files working:
		challenging-modified.pdf
		transitions_test.pdf

### radical.txt
Evaluate the potential impact of each
s t r a t e g y u s i n g t h e V i s i o n /


TEST FILES
---------
bruce.pdf for char spacing save/restore.

challenging-modified.pdf
transitions_test.pdf


Code Restructure?
-----------------
```
	type textPara struct {
		serial             int                // Sequence number for debugging.
		model.PdfRectangle                    // Bounding box.
		w, h   int
		cells []textCell
	}

	type textCell struct {
		serial             int                // Sequence number for debugging.
		model.PdfRectangle                    // Bounding box.
		eBBox              model.PdfRectangle // Extended bounding box needed to compute reading order.
		lines              []*textLine        // Paragraph text gets broken into lines.
	}
```

  x     x    x      x     x     x
  x
  x     x
  x
  x     x           x
  x
  x

1. Compute all row candidates
     alignedY  No intervening paras
2. Compute all column candidates
     alignedX  No intervening paras

Table candidate
1. Top row fully populated
2. Left column fully populated
3. All cells in table are aligned with 1 top row element and 1 left column candidate
4. Mininum number of cells must be filled

Computation time
1. Row candidates  O(N)
   Sort top to bottom, left to right
   Search
2. Column candidates O(N)
   Sort left to right, top to bottom
   Search
3. Find intersections  O(N^2)
   For each row
      Find columns that start at row -> table candiates
   Sort table candidates by w x h descending
4. Test each candidate O(N^4)
