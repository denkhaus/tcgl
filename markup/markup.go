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
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

//--------------------
// CONST
//--------------------

const RELEASE = "Tideland Common Go Library - Simple Markup Language - Release 2012-01-24"

//--------------------
// PROCESSOR
//--------------------

// Processor represents any type able to process a simple
// markup language document.
type Processor interface {
	OpenTag(tag []string)
	CloseTag(tag []string)
	Text(text string)
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

	tn := &TagNode{
		tag:      strings.Split(tmp, ":"),
		children: make([]Node, 0),
	}

	return tn
}

// AppendTag creates a new tag node, appends it as last child
// and returns it.
func (tn *TagNode) AppendTag(tag string) *TagNode {
	n := NewTagNode(tag)

	if n != nil {
		tn.children = append(tn.children, n)
	}

	return n
}

// AppendTagNode appends a tag node as last child and
// returns it.
func (tn *TagNode) AppendTagNode(n *TagNode) *TagNode {
	tn.children = append(tn.children, n)

	return n
}

// AppendText create a text node, appends it as last child
// returns it.
func (tn *TagNode) AppendText(text string) *TextNode {
	n := NewTextNode(text)

	tn.children = append(tn.children, n)

	return n
}

// AppendTaggedText creates a tag node like AppendTag() and
// for this node also a text node like AppendText(). The tag
// node will be returned.
func (tn *TagNode) AppendTaggedText(tag, text string) *TagNode {
	n := NewTagNode(tag)

	if n != nil {
		n.AppendText(text)

		tn.children = append(tn.children, n)
	}

	return n
}

// AppendTextNode appends a text node as last child and
// returns it.
func (tn *TagNode) AppendTextNode(n *TextNode) *TextNode {
	tn.children = append(tn.children, n)

	return n
}

// Len return the number of children of this node.
func (tn *TagNode) Len() int {
	return len(tn.children)
}

// ProcessWith processes the node and all chidlren recursively
// with the passed processor.
func (tn *TagNode) ProcessWith(p Processor) {
	p.OpenTag(tn.tag)

	for _, child := range tn.children {
		child.ProcessWith(p)
	}

	p.CloseTag(tn.tag)
}

// String returns the tag node as string.
func (tn *TagNode) String() string {
	buf := bytes.NewBufferString("")
	spp := NewSmlWriterProcessor(buf, true)

	tn.ProcessWith(spp)

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
	return &TextNode{text}
}

// Len returns the len of the text if the text node.
func (tn *TextNode) Len() int {
	return len(tn.text)
}

// ProcessWith processes the text node with the given
// processor.
func (tn *TextNode) ProcessWith(p Processor) {
	p.Text(tn.text)
}

// String returns the text node as string.
func (tn *TextNode) String() string {
	return tn.text
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

// Control values.
const (
	ctrlText int = iota+1
	ctrlSpace
	ctrlOpen
	ctrlClose
	ctrlEscape
	ctrlTag
	ctrlEOF
	ctrlInvalid
)

// Node read modes.
const (
	modeInit = iota
	modeTag
	modeText
)

// SmlReader is a reader creating a SML document
// out of a reader stream.
type SmlReader struct {
	reader *bufio.Reader
	index  int
	root   *TagNode
	error  error
}

// NewSmlReader creates a new reader for SML documents.
func NewSmlReader(reader io.Reader) *SmlReader {
	// Init the reader.
	sr := &SmlReader{
		reader: bufio.NewReader(reader),
		index:  -1,
	}

	node, ctrl := sr.readNode()

	switch ctrl {
	case ctrlClose:
		sr.root = node
		sr.error = nil
	case ctrlEOF:
		msg := fmt.Sprintf("eof too early at index %v", sr.index)

		sr.error = errors.New(msg)
	case ctrlInvalid:
		msg := fmt.Sprintf("invalid rune at index %v", sr.index)

		sr.error = errors.New(msg)
	}

	return sr
}

// RootTagNode returns the root tag node of the document
func (sr *SmlReader) RootTagNode() (*TagNode, error) {
	return sr.root, sr.error
}

// readNode reads the next node from the stream. This may be
// executed recursively.
func (sr *SmlReader) readNode() (*TagNode, int) {
	var node *TagNode
	var buffer *bytes.Buffer

	mode := modeInit

	for {
		rune, ctrl := sr.readRune()

		sr.index++

		switch mode {
		case modeInit:
			// Before the first opening bracket.
			switch ctrl {
			case ctrlEOF:
				return nil, ctrlEOF
			case ctrlOpen:
				mode = modeTag
				buffer = bytes.NewBufferString("")
			}
		case modeTag:
			// Reading a tag.
			switch ctrl {
			case ctrlEOF:
				return nil, ctrlEOF
			case ctrlTag:
				buffer.WriteRune(rune)
			case ctrlSpace:
				if buffer.Len() == 0 {
					return nil, ctrlInvalid
				}

				node = NewTagNode(buffer.String())
				buffer = bytes.NewBufferString("")
				mode = modeText
			case ctrlClose:
				if buffer.Len() == 0 {
					return nil, ctrlInvalid
				}

				node = NewTagNode(buffer.String())

				return node, ctrlClose
			default:
				return nil, ctrlInvalid
			}
		case modeText:
			// Reading the text including the subnodes following
			// the space after the tag or id.
			switch ctrl {
			case ctrlEOF:
				return nil, ctrlEOF
			case ctrlOpen:
				text := strings.TrimSpace(buffer.String())

				if len(text) > 0 {
					node.AppendText(text)
				}

				buffer = bytes.NewBufferString("")

				sr.reader.UnreadRune()

				subnode, subctrl := sr.readNode()

				if subctrl == ctrlClose {
					// Correct closed subnode.

					node.AppendTagNode(subnode)
				} else {
					// Error while reading the subnode.

					return nil, subctrl
				}
			case ctrlClose:
				text := strings.TrimSpace(buffer.String())

				if len(text) > 0 {
					node.AppendText(text)
				}

				return node, ctrlClose
			case ctrlEscape:
				rune, ctrl = sr.readRune()

				if ctrl == ctrlOpen || ctrl == ctrlClose || ctrl == ctrlEscape {
					buffer.WriteRune(rune)

					sr.index++
				} else {
					return nil, ctrlInvalid
				}
			default:
				buffer.WriteRune(rune)
			}
		}
	}

	return nil, ctrlEOF
}

// Reads one rune of the reader.
func (sr *SmlReader) readRune() (r rune, ctrl int) {
	var size int

	r, size, sr.error = sr.reader.ReadRune()

	switch {
	case size == 0:
		return r, ctrlEOF
	case r == '{':
		return r, ctrlOpen
	case r == '}':
		return r, ctrlClose
	case r == '^':
		return r, ctrlEscape
	case r >= 'a' && r <= 'z':
		return r, ctrlTag
	case r >= 'A' && r <= 'Z':
		return r, ctrlTag
	case r >= '0' && r <= '9':
		return r, ctrlTag
	case r == '-':
		return r, ctrlTag
	case r == ':':
		return r, ctrlTag
	case unicode.IsSpace(r):
		return r, ctrlSpace
	}

	return r, ctrlText
}

//--------------------
// SML WRITER PROCESSOR
//--------------------

// SmlWriterProcessor writes a SML document to a writer.
type SmlWriterProcessor struct {
	writer      *bufio.Writer
	prettyPrint bool
	indentLevel int
}

// NewSmlWriterProcessor creates a new writer for a SML document.
func NewSmlWriterProcessor(writer io.Writer, prettyPrint bool) *SmlWriterProcessor {
	swp := &SmlWriterProcessor{
		writer:      bufio.NewWriter(writer),
		prettyPrint: prettyPrint,
		indentLevel: 0,
	}

	return swp
}

// OpenTag writes the opening of a tag.
func (swp *SmlWriterProcessor) OpenTag(tag []string) {
	swp.writeIndent(true)

	swp.writer.WriteString("{")
	swp.writer.WriteString(strings.Join(tag, ":"))
}

// CloseTag writes the closing of a tag.
func (swp *SmlWriterProcessor) CloseTag(tag []string) {
	swp.writer.WriteString("}")

	if swp.prettyPrint {
		swp.indentLevel--
	}

	swp.writer.Flush()
}

// Text writes a text with an encoding of special runes.
func (swp *SmlWriterProcessor) Text(text string) {
	ta := strings.Replace(text, "^", "^^", -1)
	tb := strings.Replace(ta, "{", "^{", -1)
	tc := strings.Replace(tb, "}", "^}", -1)

	swp.writeIndent(false)

	swp.writer.WriteString(tc)
}

// writeIndent writes an indentation of wanted.
func (swp *SmlWriterProcessor) writeIndent(increase bool) {
	if swp.prettyPrint {
		if swp.indentLevel > 0 {
			swp.writer.WriteString("\n")
		}

		for i := 0; i < swp.indentLevel; i++ {
			swp.writer.WriteString("\t")
		}

		if increase {
			swp.indentLevel++
		}
	} else {
		swp.writer.WriteString(" ")
	}
}

// EOF
