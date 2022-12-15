package lang

type TextKind string

const (
	PlainText   TextKind = "plain"
	ItalicText  TextKind = "italic"
	LinkText    TextKind = "link"
	DocLinkText TextKind = "docLink"
)

type TextBlock struct {
	kind  TextKind
	text  string
	path  string
	href  string
	inner *Text
}

func NewTextBlock(kind TextKind, text string) *TextBlock {
	return &TextBlock{kind: kind, text: text}
}

func NewLinkTextBlock(kind TextKind, inner []*TextBlock, path string, href string) *TextBlock {
	return &TextBlock{kind: kind, inner: NewText(inner), path: path, href: href}
}

func (t *TextBlock) Kind() TextKind {
	return t.kind
}

func (t *TextBlock) Text() string {
	return t.text
}

func (t *TextBlock) Path() string {
	return t.path
}

func (t *TextBlock) Href() string {
	return t.href
}

func (t *TextBlock) Inner() *Text {
	return t.inner
}

type Text struct {
	textBlocks []*TextBlock
}

func NewText(textBlocks []*TextBlock) *Text {
	return &Text{textBlocks}
}

func (t *Text) TextBlocks() []*TextBlock {
	return t.textBlocks
}
