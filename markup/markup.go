// Tideland Common Go Library - Simple Markup Language
//
// Copyright (C) 2009-2012 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed 
// by the new BSD license.

package markup

//--------------------
// IMPORTS
//--------------------

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

//--------------------
// PROCESSOR
//--------------------

// Processor represents any type able to process a simple
// markup language document.
type Processor interface {
	OpenTag(tag []string)
	CloseTag(tag []string)
	Text(text string)
	Raw(raw string)
}

//--------------------
// NODE
//--------------------

// Node represents the common interface of all nodes (tags and text).
type Node interface {
	Len() int
	ProcessWith(p Processor)
}

//--------------------
// TAG NODE
//--------------------

// TagNode represents a node with one multipart tag and zero to many
// children nodes.
type TagNode struct {
	tag      []string
	children []Node
}

// NewTagNode create a new tag node with the given tag. It will
// be lower-cased and validated. Each tag part only contains the
// chars 'a' to 'z', '0' to '9' and '-'. The separator is a colon.
func NewTagNode(tag string) *TagNode {
	tmp := strings.ToLower(tag)
	if !validIdentifier(tmp) {
		return nil
	}
	n := &TagNode{
		tag:      strings.Split(tmp, ":"),
		children: make([]Node, 0),
	}
	return n
}

// AppendTagNode creates a new tag node, appends it as last child
// and returns it.
func (t *TagNode) AppendTagNode(tag string) Node {
	n := NewTagNode(tag)
	if n != nil {
		t.AppendNode(n)
	}
	return n
}

// AppendTextNode create a text node, appends it as last child
// and returns it.
func (t *TagNode) AppendTextNode(text string) Node {
	return t.AppendNode(NewTextNode(text))
}

// AppendTaggedTextNode creates a tag node like AppendTagNode() and
// for this node also a text node like AppendTextNode(). The tag
// node will be returned.
func (t *TagNode) AppendTaggedTextNode(tag, text string) Node {
	n := t.AppendTagNode(tag).(*TagNode)
	if n != nil {
		n.AppendTextNode(text)
	}
	return n
}

// AppendRawNode create a raw node, appends it as last child
// and returns it.
func (t *TagNode) AppendRawNode(raw string) Node {
	return t.AppendNode(NewRawNode(raw))
}

// AppendNode appends a node
func (t *TagNode) AppendNode(n Node) Node {
	t.children = append(t.children, n)
	return n
}

// Tag returns the tag parts joined by a colon.
func (t *TagNode) Tag() string {
	return strings.Join(t.tag, ":")
}

// Len return the number of children of this node.
func (t *TagNode) Len() int {
	return 1 + len(t.children)
}

// ProcessWith processes the node and all chidlren recursively
// with the passed processor.
func (t *TagNode) ProcessWith(p Processor) {
	p.OpenTag(t.tag)
	for _, child := range t.children {
		child.ProcessWith(p)
	}
	p.CloseTag(t.tag)
}

// String returns the tag node as string.
func (t *TagNode) String() string {
	buf := bytes.NewBufferString("")
	spp := NewSMLWriterProcessor(buf, true)
	t.ProcessWith(spp)
	return buf.String()
}

//--------------------
// TEXT NODE
//--------------------

// TextNode is a node containing some text.
type TextNode struct {
	text string
}

// NewTextNode creates a new text node.
func NewTextNode(text string) *TextNode {
	return &TextNode{strings.TrimSpace(text)}
}

// Len returns the len of the text in the text node.
func (t *TextNode) Len() int {
	return len(t.text)
}

// ProcessWith processes the text node with the given
// processor.
func (t *TextNode) ProcessWith(p Processor) {
	p.Text(t.text)
}

// String returns the text node as string.
func (t *TextNode) String() string {
	return t.text
}

//--------------------
// RAW NODE
//--------------------

// RawNode is a node containing some raw data.
type RawNode struct {
	raw string
}

// NewRawNode creates a new raw node.
func NewRawNode(raw string) *RawNode {
	return &RawNode{strings.TrimSpace(raw)}
}

// Len returns the len of the data in the raw node.
func (r *RawNode) Len() int {
	return len(r.raw)
}

// ProcessWith processes the raw node with the given
// processor.
func (r *RawNode) ProcessWith(p Processor) {
	p.Raw(r.raw)
}

// String returns the raw node as string.
func (r *RawNode) String() string {
	return r.raw
}

//--------------------
// PRIVATE FUNCTIONS
//--------------------

// validIdentifier checks if an identifier is valid. Only
// the chars 'a' to 'z', '0' to '9', '-' and ':' are
// accepted.
func validIdentifier(id string) bool {
	for _, c := range id {
		if c < 'a' || c > 'z' {
			if c < '0' || c > '9' {
				if c != '-' && c != ':' {
					return false
				}
			}
		}
	}
	return true
}

//--------------------
// SML READER
//--------------------

// Rune classes.
const (
	rcText int = iota + 1
	rcSpace
	rcOpen
	rcClose
	rcEscape
	rcExclamation
	rcTag
	rcEOF
	rcInvalid
)

// ReadSML parses a SML document and returns it as 
// node structure.
func ReadSML(reader io.Reader) (*TagNode, error) {
	s := &smlReader{
		reader: bufio.NewReader(reader),
		index:  -1,
	}
	err := s.readPreliminary()
	if err != nil {
		return nil, err
	}
	return s.readTagNode()
}

// smlReader is used by ReadSML to parse a SML document
// and return it as node structure.
type smlReader struct {
	reader *bufio.Reader
	index  int
}

// readPreliminary reads the content before the first node.
func (s *smlReader) readPreliminary() error {
	for {
		_, rc, err := s.readRune()
		switch {
		case err != nil:
			return err
		case rc == rcEOF:
			return fmt.Errorf("unexpected end of file while reading preliminary")
		case rc == rcOpen:
			return nil
		}
	}
	// Unreachable.
	panic("unreachable")
}

// readNode reads the next tag node.
func (s *smlReader) readTagNode() (*TagNode, error) {
	tag, rc, err := s.readTag()
	if err != nil {
		return nil, err
	}
	node := NewTagNode(tag)
	// Read children.
	if rc != rcClose {
		err = s.readTagChildren(node)
		if err != nil {
			return nil, err
		}
	}
	return node, nil
}

// readTag reads the tag of a node. It als returns the class of the next rune.
func (s *smlReader) readTag() (string, int, error) {
	var buf bytes.Buffer
	for {
		r, rc, err := s.readRune()
		switch {
		case err != nil:
			return "", 0, err
		case rc == rcEOF:
			return "", 0, fmt.Errorf("unexpected end of file while reading a tag")
		case rc == rcTag:
			buf.WriteRune(r)
		case rc == rcSpace || rc == rcClose:
			return buf.String(), rc, nil
		default:
			return "", 0, fmt.Errorf("invalid tag rune at index %d", s.index)
		}
	}
	// Unreachable.
	panic("unreachable")
}

// readChildren reads the children of passed parent tag node.
func (s *smlReader) readTagChildren(p *TagNode) error {
	for {
		_, rc, err := s.readRune()
		switch {
		case err != nil:
			return err
		case rc == rcEOF:
			return fmt.Errorf("unexpected end of file while reading children")
		case rc == rcClose:
			return nil
		case rc == rcOpen:
			node, err := s.readTagOrRawNode()
			if err != nil {
				return err
			}
			if node.Len() > 0 {
				p.AppendNode(node)
			}
		default:
			s.index--
			s.reader.UnreadRune()
			node, err := s.readTextNode()
			if err != nil {
				return err
			}
			if node.Len() > 0 {
				p.AppendNode(node)
			}
		}
	}
	// Unreachable.
	panic("unreachable")
}

// readTagOrRawNode checks if the opening is for a tag node or
// for a raw node and starts the reading of it.
func (s *smlReader) readTagOrRawNode() (Node, error) {
	_, rc, err := s.readRune()
	switch {
	case err != nil:
		return nil, err
	case rc == rcEOF:
		return nil, fmt.Errorf("unexpected end of file while reading a tag or raw node")
	case rc == rcTag:
		s.index--
		s.reader.UnreadRune()
		return s.readTagNode()
	case rc == rcExclamation:
		return s.readRawNode()
	}
	return nil, fmt.Errorf("invalid rune after opening at index %d", s.index)
}

// readRawNode reads a raw node.
func (s *smlReader) readRawNode() (*RawNode, error) {
	var buf bytes.Buffer
	for {
		r, rc, err := s.readRune()
		switch {
		case err != nil:
			return nil, err
		case rc == rcEOF:
			return nil, fmt.Errorf("unexpected end of file while reading a raw node")
		case rc == rcExclamation:
			r, rc, err = s.readRune()
			switch {
			case err != nil:
				return nil, err
			case rc == rcEOF:
				return nil, fmt.Errorf("unexpected end of file while reading a raw node")
			case rc == rcClose:
				return NewRawNode(buf.String()), nil
			}
			buf.WriteRune('!')
			buf.WriteRune(r)
		default:
			buf.WriteRune(r)
		}
	}
	// Unreachable.
	panic("unreachable")
}

// readTextNode reads a text node.
func (s *smlReader) readTextNode() (*TextNode, error) {
	var buf bytes.Buffer
	for {
		r, rc, err := s.readRune()
		switch {
		case err != nil:
			return nil, err
		case rc == rcEOF:
			return nil, fmt.Errorf("unexpected end of file while reading a text node")
		case rc == rcOpen || rc == rcClose:
			s.index--
			s.reader.UnreadRune()
			return NewTextNode(buf.String()), nil
		case rc == rcEscape:
			r, rc, err = s.readRune()
			switch {
			case err != nil:
				return nil, err
			case rc == rcEOF:
				return nil, fmt.Errorf("unexpected end of file while reading a text node")
			case rc == rcOpen || rc == rcClose || rc == rcEscape:
				buf.WriteRune(r)
			default:
				return nil, fmt.Errorf("invalid rune after escape at index %d", s.index)
			}
		default:
			buf.WriteRune(r)
		}
	}
	// Unreachable.
	panic("unreachable")
}

// Reads one rune of the reader.
func (s *smlReader) readRune() (r rune, rc int, err error) {
	var size int
	s.index++
	r, size, err = s.reader.ReadRune()
	if err != nil {
		return 0, 0, err
	}
	switch {
	case size == 0:
		rc = rcEOF
	case r == '{':
		rc = rcOpen
	case r == '}':
		rc = rcClose
	case r == '^':
		rc = rcEscape
	case r == '!':
		rc = rcExclamation
	case r >= 'a' && r <= 'z':
		rc = rcTag
	case r >= 'A' && r <= 'Z':
		rc = rcTag
	case r >= '0' && r <= '9':
		rc = rcTag
	case r == '-' || r == ':':
		rc = rcTag
	case unicode.IsSpace(r):
		rc = rcSpace
	default:
		rc = rcText
	}
	return
}

//--------------------
// SML WRITER PROCESSOR
//--------------------

// SMLWriterProcessor writes a SML document to a writer.
type SMLWriterProcessor struct {
	writer      *bufio.Writer
	prettyPrint bool
	indentLevel int
}

// NewSMLWriterProcessor creates a new writer for a SML document.
func NewSMLWriterProcessor(writer io.Writer, prettyPrint bool) *SMLWriterProcessor {
	return &SMLWriterProcessor{
		writer:      bufio.NewWriter(writer),
		prettyPrint: prettyPrint,
		indentLevel: 0,
	}
}

// OpenTag writes the opening of a tag.
func (p *SMLWriterProcessor) OpenTag(tag []string) {
	p.writeIndent(true)
	p.writer.WriteString("{")
	p.writer.WriteString(strings.Join(tag, ":"))
}

// CloseTag writes the closing of a tag.
func (p *SMLWriterProcessor) CloseTag(tag []string) {
	p.writer.WriteString("}")
	if p.prettyPrint {
		p.indentLevel--
	}
	p.writer.Flush()
}

// Text writes a text with an encoding of special runes.
func (p *SMLWriterProcessor) Text(text string) {
	t := strings.Replace(text, "^", "^^", -1)
	t = strings.Replace(t, "{", "^{", -1)
	t = strings.Replace(t, "}", "^}", -1)
	p.writeIndent(false)
	p.writer.WriteString(t)
}

// Raw write a raw data without any encoding.
func (p *SMLWriterProcessor) Raw(raw string) {
	p.writeIndent(false)
	p.writer.WriteString("{! ")
	p.writer.WriteString(raw)
	p.writer.WriteString(" !}")
}

// writeIndent writes an indentation of wanted.
func (p *SMLWriterProcessor) writeIndent(increase bool) {
	if p.prettyPrint {
		if p.indentLevel > 0 {
			p.writer.WriteString("\n")
		}
		for i := 0; i < p.indentLevel; i++ {
			p.writer.WriteString("\t")
		}
		if increase {
			p.indentLevel++
		}
	} else {
		p.writer.WriteString(" ")
	}
}

// EOF
