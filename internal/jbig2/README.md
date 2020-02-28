## UNIPDF JBIG2

### Brief Description

JBIG2 is the standard for bi-level image compression, developed by the Joint Bi-level Image Experts Group.
It is designed to compress black and white images in both lossless and lossy modes with better performance than
traditional JBIG and Fax Group 4 standards.

### Performance
The file size of a typical scanned document at 300dpi for a TIFF is around 75KB-125 KB per image.
The same document encoded using JBIG2 would be about 5 up to 10 times smaller (10KB - 15KB per image).

### Unipdf JBIG2Encoder
JBIG2 standard allows to encode bi-level images in two modes:

- lossless  - the image quality is the same as original, no data is lost
- lossy - better compression ration, but some image parts might be lost

##### Generic region encoding
Unipdf library allows to encode black and white images losslessly by providing Generic method.
The encoder takes whole image as a generic region and encodes it using arithmetic coder.
It allows to reduce the file size by encoding a line duplicates using a single bit. This is used by
setting `DuplicateLinesRemoval`. The more lines are duplicated the better the compression rate. 

This method is relatively fast with a basic compression.

##### (Upcoming) Classified - component encoding
Unipdf is working on a classified component, lossy encoding method. The encoder reads and scans all pages for provided document.
Their content is being decomposed into symbols and matched for similar occurrences. The symbols are stored in a
Symbol Dictionary segment using arithmetic coder.
As the next step, encoder creates Text Region Segments that contains positions of provided symbols on the page bitmap.


#### Options

### Examples

### Unipdf JBIG2Decoder
Unipdf library allows to decode JBIG2 encoded files and byte streams.

#### Decoder Examples


### Acknowledges