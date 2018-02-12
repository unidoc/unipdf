package creator

// PageBreak represents a page break for a chapter.
type PageBreak struct {
}

// NewPageBreak create a new page break.
func NewPageBreak() *PageBreak {
	return &PageBreak{}
}

// GeneratePageBlocks generates a page break block.
func (p *PageBreak) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	blocks := []*Block{
		NewBlock(ctx.Width, ctx.PageHeight),
	}
	ctx.Y = ctx.PageHeight
	return blocks, ctx, nil
}
