/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

// The table of contents has overview over chapters and subchapters.
type TableOfContents struct {
	entries []tableOfContentsEntry
}

// Make a new table of contents.
func newTableOfContents() *TableOfContents {
	toc := TableOfContents{}
	toc.entries = []tableOfContentsEntry{}
	return &toc
}

// Get table of content entries.
func (toc *TableOfContents) Entries() []tableOfContentsEntry {
	return toc.entries
}

// Add a TOC entry.
func (toc *TableOfContents) add(title string, chapter, subchapter, pageNum int) {
	entry := tableOfContentsEntry{}
	entry.Title = title
	entry.Chapter = chapter
	entry.Subchapter = subchapter
	entry.PageNumber = pageNum

	toc.entries = append(toc.entries, entry)
}

// Each TOC entry has title, chapter number, sub chapter (0 if chapter) and the page number.
type tableOfContentsEntry struct {
	Title      string
	Chapter    int
	Subchapter int // 0 if chapter
	PageNumber int // Page number
}
