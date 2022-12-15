package lang

import (
	"go/doc"
	"go/doc/comment"
	"regexp"
	"strings"
)

// Doc provides access to the documentation comment contents for a package or
// symbol in a structured form.
type Doc struct {
	cfg    *Config
	blocks []*Block
}

var (
	multilineRegex      = regexp.MustCompile("\n(?:[\t\f ]*\n)+")
	headerRegex         = regexp.MustCompile(`^[A-Z][^!:;,{}\[\]<>.?]*\n(?:[\t\f ]*\n)`)
	spaceCodeBlockRegex = regexp.MustCompile(`^(?:(?:(?:(?:  ).*[^\s]+.*)|[\t\f ]*)\n)+`)
	tabCodeBlockRegex   = regexp.MustCompile(`^(?:(?:(?:\t.*[^\s]+.*)|[\t\f ]*)\n)+`)
	blankLineRegex      = regexp.MustCompile(`^[\t\f ]*\n`)
)

// NewDoc initializes a Doc struct from the provided raw documentation text and
// with headers rendered by default at the heading level provided. Documentation
// is separated into block level elements using the standard rules from golang's
// documentation conventions.
func NewDoc(cfg *Config, currentPkg *doc.Package, allPackages []*doc.Package, text string) *Doc {
	doc := currentPkg.Parser().Parse(text)
	module, _ := getModule(cfg.WorkDir)
	var blocks []*Block
	for _, block := range doc.Content {
		switch block := block.(type) {
		case *comment.Code:
			text := NewTextBlock(PlainText, block.Text)
			blocks = append(blocks, NewBlock(cfg.Inc(0), CodeBlock, NewText([]*TextBlock{text})))
		case *comment.Heading:
			text := textBlocks(block.Text, module, currentPkg, allPackages)
			blocks = append(blocks, NewBlock(cfg.Inc(0), HeaderBlock, NewText(text)))
		case *comment.List:
			text := []*TextBlock{}
			if block.BlankBefore() {
				text = append(text, NewTextBlock(PlainText, "\n"))
			}
			for _, item := range block.Items {
				text = append(text, NewTextBlock(PlainText, item.Number))
				commentText := []comment.Text{}
				for _, t := range item.Content {
					p := t.(*comment.Paragraph)
					commentText = append(commentText, p.Text...)
				}
				text = append(text, textBlocks(commentText, module, currentPkg, allPackages)...)
				blocks = append(blocks, NewBlock(cfg.Inc(0), ListBlock, NewText(text)))
			}
			blocks = append(blocks, NewBlock(cfg.Inc(0), ListBlock, NewText(text)))
		case *comment.Paragraph:
			text := textBlocks(block.Text, module, currentPkg, allPackages)
			blocks = append(blocks, NewBlock(cfg.Inc(0), ParagraphBlock, NewText(text)))
		}
	}

	return &Doc{cfg, blocks}
}

// Level provides the default level that headers within the documentation should
// be rendered
func (d *Doc) Level() int {
	return d.cfg.Level
}

// Blocks holds the list of block elements that makes up the documentation
// contents.
func (d *Doc) Blocks() []*Block {
	return d.blocks
}

func getImportType(link *comment.DocLink, currentPkg *doc.Package, allPackages []*doc.Package) string {
	if link.ImportPath == "" {
		// If no import path, this is a link to the current package
		return getIdentifierType(link, currentPkg)
	}
	for _, pkg := range allPackages {
		if pkg.ImportPath == link.ImportPath && pkg.Parser().LookupSym(link.Recv, link.Name) {
			return getIdentifierType(link, pkg)
		}
	}
	return ""
}

func getIdentifierType(link *comment.DocLink, pkg *doc.Package) string {
	for _, constIdent := range pkg.Consts {
		for _, constName := range constIdent.Names {
			if link.Name == constName {
				return "const"
			}
		}
	}
	for _, constIdent := range pkg.Vars {
		for _, constName := range constIdent.Names {
			if link.Name == constName {
				return "var"
			}
		}
	}
	for _, funcIdent := range pkg.Funcs {
		if funcIdent.Recv == link.Recv && funcIdent.Name == link.Name {
			return "func"
		}
	}
	for _, typeIdent := range pkg.Types {
		if typeIdent.Name == link.Name {
			return "type"
		}
	}

	return ""
}

func textBlocks(text []comment.Text, module string, currentPkg *doc.Package, allPackages []*doc.Package) []*TextBlock {
	blocks := []*TextBlock{}
	for _, line := range text {
		switch line := line.(type) {
		case comment.Plain:
			blocks = append(blocks, NewTextBlock(PlainText, string(line)))
		case comment.Italic:
			blocks = append(blocks, NewTextBlock(ItalicText, string(line)))
		case *comment.Link:
			blocks = append(blocks, NewLinkTextBlock(LinkText, textBlocks(line.Text, module, currentPkg, allPackages), "", line.URL))
		case *comment.DocLink:
			importType := getImportType(line, currentPkg, allPackages)
			replaceStr := line.ImportPath
			if line.ImportPath == "" {
				// If no import path, this is a link to the current package
				replaceStr = currentPkg.ImportPath
			}
			path := strings.Replace(replaceStr, module, "", 1)
			href := importType
			if line.Recv != "" {
				href += " " + line.Recv
			}
			href += " " + line.Name

			blocks = append(blocks, NewLinkTextBlock(DocLinkText, textBlocks(line.Text, module, currentPkg, allPackages), path, href))
		}
	}
	return blocks
}
