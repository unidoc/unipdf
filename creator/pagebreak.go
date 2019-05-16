package creator

// PageBreak represents a page break for a chapter.
type PageBreak struct {
}

// newPageBreak create a new page break.
func newPageBreak() *PageBreak {
	return &PageBreak{}
}

// GeneratePageBlocks generates a page break block.
func (p *PageBreak) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	// Return two empty blocks.  First one simply means that there is nothing more to add at the current page.
	// The second one starts a new page.
	blocks := []*Block{
		NewBlock(ctx.PageWidth, ctx.PageHeight-ctx.Y),
		NewBlock(ctx.PageWidth, ctx.PageHeight),
	}

	// New Page. Place context in upper left corner (with margins).
	ctx.Page++
	newContext := ctx
	newContext.Y = ctx.Margins.top
	newContext.X = ctx.Margins.left
	newContext.Height = ctx.PageHeight - ctx.Margins.top - ctx.Margins.bottom
	newContext.Width = ctx.PageWidth - ctx.Margins.left - ctx.Margins.right
	ctx = newContext

	return blocks, ctx, nil
}
