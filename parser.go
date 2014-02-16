package goose

import (
	"code.google.com/p/go.net/html"
	"github.com/PuerkitoBio/goquery"
)

type parser struct{}

func NewParser() *parser {
	return &parser{}
}

func (this *parser) dropTag(selection *goquery.Selection) {
	selection.Each(func(i int, s *goquery.Selection) {
		node := s.Get(0)
		node.Data = s.Text()
		node.Type = html.TextNode
	})
}

func (this *parser) indexOfAttribute(selection *goquery.Selection, attr string) int {
	node := selection.Get(0)
	for i, a := range node.Attr {
		if a.Key == attr {
			return i
		}
	}
	return -1
}

func (this *parser) delAttr(selection *goquery.Selection, attr string) {
	idx := this.indexOfAttribute(selection, attr)
	if idx > -1 {
		node := selection.Get(0)
		node.Attr = append(node.Attr[:idx], node.Attr[idx+1:]...)
	}
}

func (this *parser) getElementsByTags(div *goquery.Selection, tags []string) *goquery.Selection {
	selection := new(goquery.Selection)
	for _, tag := range tags {
		selections := div.Find(tag)
		if selections != nil {
			selection = selection.Union(selections)
		}
	}
	return selection
}

func (this *parser) clear(selection *goquery.Selection) {
	selection.Nodes = make([]*html.Node, 0)
}

func (this *parser) removeNode(selection *goquery.Selection) {
	if selection != nil {
		node := selection.Get(0)
		if node != nil && node.Parent != nil {
			node.Parent.RemoveChild(node)
		}
	}
}

func (this *parser) setAttr(selection *goquery.Selection, attr string, value string) {
	node := selection.Get(0)

	for _, a := range node.Attr {
		if a.Key == attr {
			a.Val = value
			return
		}
	}
	attrs := make([]html.Attribute, len(node.Attr)+1)
	for i, a := range node.Attr {
		attrs[i+1] = a
	}
	newAttr := new(html.Attribute)
	newAttr.Key = attr
	newAttr.Val = value
	attrs[0] = *newAttr
	node.Attr = attrs
}
