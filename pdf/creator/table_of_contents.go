/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

// TableOfContents provides an overview over chapters and subchapters when creating a document with Creator.
type TableOfContents struct {
	entries []TableOfContentsEntry
}

// Make a new table of contents.
func newTableOfContents() *TableOfContents {
	toc := TableOfContents{}
	toc.entries = []TableOfContentsEntry{}
	return &toc
}

// Entries returns the table of content entries.
func (toc *TableOfContents) Entries() []TableOfContentsEntry {
	return toc.entries
}

// Add a TOC entry.
func (toc *TableOfContents) add(title string, chapter, subchapter, pageNum int) {
	entry := TableOfContentsEntry{}
	entry.Title = title
	entry.Chapter = chapter
	entry.Subchapter = subchapter
	entry.PageNumber = pageNum

	toc.entries = append(toc.entries, entry)
}

// TableOfContentsEntry defines a single entry in the TableOfContents.
// Each entry has a title, chapter number, sub chapter (0 if chapter) and the page number.
type TableOfContentsEntry struct {
	Title      string
	Chapter    int
	Subchapter int // 0 if chapter
	PageNumber int // Page number
}
