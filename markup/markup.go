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

// Processor represents any type able to process
// a node structure.
type Processor interface {
	OpenTag(tag []string) error
	CloseTag(tag []string) error
	Text(text string) error
	Raw(raw string) error
}

//--------------------
// BUILDER
//--------------------

// Builder represents any type able to build
// something out of a parsed SML document.
type Builder interface {
	BeginTagNode(tag string) error
	EndTagNode() error
	AppendTextNode(text string) error
	AppendRawNode(raw string) error
}

//--------------------
// NODE
//--------------------

// Node represents the common interface of all nodes (tags and text).
type Node interface {
	Len() int
	ProcessWith(p Processor) error
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
func (t *TagNode) ProcessWith(p Processor) error {
	if err := p.OpenTag(t.tag); err != nil {
		return err
	}
	for _, child := range t.children {
		if err := child.ProcessWith(p); err != nil {
			return err
		}
	}
	return p.CloseTag(t.tag)
}

// String returns the tag node as string.
func (t *TagNode) String() string {
	var buf bytes.Buffer
	WriteSML(t, &buf, true)
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
func (t *TextNode) ProcessWith(p Processor) error {
	return p.Text(t.text)
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
func (r *RawNode) ProcessWith(p Processor) error {
	return p.Raw(r.raw)
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
// NODE BUILDER
//--------------------

// NodeBuilder creates a node structure when a SML
// document is read.
type NodeBuilder struct {
	root  *TagNode
	stack []*TagNode
}

// NewNodeBuilder return a new nnode builder.
func NewNodeBuilder() *NodeBuilder {
	return &NodeBuilder{nil, make([]*TagNode, 0)}
}

// Root returns the root node of the read document.
func (n *NodeBuilder) Root() *TagNode {
	return n.root
}

// BeginTagNode opens a new tag node.
func (n *NodeBuilder) BeginTagNode(tag string) error {
	n.stack = append(n.stack, NewTagNode(tag))
	return nil
}

// EndTagNode closes a new tag node.
func (n *NodeBuilder) EndTagNode() error {
	l := len(n.stack)
	if l > 1 {
		n.stack[l-2].AppendNode(n.stack[l-1])
		n.stack = n.stack[:l-1]
	} else {
		n.root = n.stack[0]
		n.stack = nil
	}
	return nil
}

// AppendTextNode appends a text node to the current open node.
func (n *NodeBuilder) AppendTextNode(text string) error {
	t := strings.TrimSpace(text)
	if len(t) > 0 {
		n.stack[len(n.stack)-1].AppendTextNode(t)
	}
	return nil
}

// AppendRawNode appends a raw node to the current open node.
func (n *NodeBuilder) AppendRawNode(raw string) error {
	r := strings.TrimSpace(raw)
	if len(r) > 0 {
		n.stack[len(n.stack)-1].AppendRawNode(r)
	}
	return nil
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

// ReadSML parses a SML document and uses the passed builder.
func ReadSML(reader io.Reader, builder Builder) error {
	s := &smlReader{
		reader:  bufio.NewReader(reader),
		builder: builder,
		index:   -1,
	}
	if err := s.readPreliminary(); err != nil {
		return err
	}
	return s.readTagNode()
}

// smlReader is used by ReadSML to parse a SML document
// and return it as node structure.
type smlReader struct {
	reader  *bufio.Reader
	builder Builder
	index   int
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
func (s *smlReader) readTagNode() error {
	tag, rc, err := s.readTag()
	if err != nil {
		return err
	}
	if err = s.builder.BeginTagNode(tag); err != nil {
		return err
	}
	// Read children.
	if rc != rcClose {
		if err = s.readTagChildren(); err != nil {
			return err
		}
	}
	return s.builder.EndTagNode()
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
func (s *smlReader) readTagChildren() error {
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
			if err = s.readTagOrRawNode(); err != nil {
				return err
			}
		default:
			s.index--
			s.reader.UnreadRune()
			if err = s.readTextNode(); err != nil {
				return err
			}
		}
	}
	// Unreachable.
	panic("unreachable")
}

// readTagOrRawNode checks if the opening is for a tag node or
// for a raw node and starts the reading of it.
func (s *smlReader) readTagOrRawNode() error {
	_, rc, err := s.readRune()
	switch {
	case err != nil:
		return err
	case rc == rcEOF:
		return fmt.Errorf("unexpected end of file while reading a tag or raw node")
	case rc == rcTag:
		s.index--
		s.reader.UnreadRune()
		return s.readTagNode()
	case rc == rcExclamation:
		return s.readRawNode()
	}
	return fmt.Errorf("invalid rune after opening at index %d", s.index)
}

// readRawNode reads a raw node.
func (s *smlReader) readRawNode() error {
	var buf bytes.Buffer
	for {
		r, rc, err := s.readRune()
		switch {
		case err != nil:
			return err
		case rc == rcEOF:
			return fmt.Errorf("unexpected end of file while reading a raw node")
		case rc == rcExclamation:
			r, rc, err = s.readRune()
			switch {
			case err != nil:
				return err
			case rc == rcEOF:
				return fmt.Errorf("unexpected end of file while reading a raw node")
			case rc == rcClose:
				return s.builder.AppendRawNode(buf.String())
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
func (s *smlReader) readTextNode() error {
	var buf bytes.Buffer
	for {
		r, rc, err := s.readRune()
		switch {
		case err != nil:
			return err
		case rc == rcEOF:
			return fmt.Errorf("unexpected end of file while reading a text node")
		case rc == rcOpen || rc == rcClose:
			s.index--
			s.reader.UnreadRune()
			return s.builder.AppendTextNode(buf.String())
		case rc == rcEscape:
			r, rc, err = s.readRune()
			switch {
			case err != nil:
				return err
			case rc == rcEOF:
				return fmt.Errorf("unexpected end of file while reading a text node")
			case rc == rcOpen || rc == rcClose || rc == rcEscape:
				buf.WriteRune(r)
			default:
				return fmt.Errorf("invalid rune after escape at index %d", s.index)
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

// WriteSML writes a node structure as SML document.
func WriteSML(root *TagNode, writer io.Writer, prettyPrint bool) error {
	return root.ProcessWith(NewSMLWriter(writer, prettyPrint))
}

// smlWriter writes a SML document to a writer.
type smlWriter struct {
	writer      *bufio.Writer
	prettyPrint bool
	indentLevel int
}

// NewSMLWriter creates a new writer for a SML document.
func NewSMLWriter(writer io.Writer, prettyPrint bool) Processor {
	return &smlWriter{
		writer:      bufio.NewWriter(writer),
		prettyPrint: prettyPrint,
		indentLevel: 0,
	}
}

// OpenTag writes the opening of a tag.
func (s *smlWriter) OpenTag(tag []string) error {
	s.writeIndent(true)
	s.writer.WriteString("{")
	s.writer.WriteString(strings.Join(tag, ":"))
	return nil
}

// CloseTag writes the closing of a tag.
func (s *smlWriter) CloseTag(tag []string) error {
	s.writer.WriteString("}")
	if s.prettyPrint {
		s.indentLevel--
	}
	return s.writer.Flush()
}

// Text writes a text with an encoding of special runes.
func (s *smlWriter) Text(text string) error {
	t := strings.Replace(text, "^", "^^", -1)
	t = strings.Replace(t, "{", "^{", -1)
	t = strings.Replace(t, "}", "^}", -1)
	s.writeIndent(false)
	s.writer.WriteString(t)
	return nil
}

// Raw write a raw data without any encoding.
func (s *smlWriter) Raw(raw string) error {
	s.writeIndent(false)
	s.writer.WriteString("{! ")
	s.writer.WriteString(raw)
	s.writer.WriteString(" !}")
	return nil
}

// writeIndent writes an indentation of wanted.
func (s *smlWriter) writeIndent(increase bool) {
	if s.prettyPrint {
		if s.indentLevel > 0 {
			s.writer.WriteString("\n")
		}
		for i := 0; i < s.indentLevel; i++ {
			s.writer.WriteString("\t")
		}
		if increase {
			s.indentLevel++
		}
	} else {
		s.writer.WriteString(" ")
	}
}

// EOF
