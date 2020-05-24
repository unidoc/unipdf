There are two directions

- *reading*
- *depth*

In English text,
- the *reading* direction is left to right, increasing X in the PDF coordinate system.
- the *depth* directon is top to bottom, decreasing Y in the PDF coordinate system.

We define *depth* as distance from the bottom of a word's bounding box from the top of the page.
depth := pageSize.Ury - r.Lly
