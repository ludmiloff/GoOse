/*
This is a golang port of "Goose" originaly licensed to Gravity.com
under one or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.

Golang port was written by Antonio Linari

Gravity.com licenses this file
to you under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package goose

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/set"
	"strconv"
	"strings"
)

type VideoExtractor struct {
	article    *Article
	config     configuration
	candidates *set.Set
	movies     *set.Set
}

type video struct {
	embedType string
	provider  string
	width     int
	height    int
	embedCode string
	src       string
}

func NewVideoExtractor() VideoExtractor {
	return VideoExtractor{
		candidates: set.New(),
		movies:     set.New(),
	}
}

var videoTags = [4]string{"iframe", "embed", "object", "video"}
var videoProviders = [4]string{"youtube", "vimeo", "dailymotion", "kewego"}

func (ve *VideoExtractor) getEmbedCode(node *goquery.Selection) string {
	return node.Text()
}

func (ve *VideoExtractor) getWidth(node *goquery.Selection) int {
	value, exists := node.Attr("width")
	if exists {
		nvalue, _ := strconv.Atoi(value)
		return nvalue
	}
	return 0
}

func (ve *VideoExtractor) getHeight(node *goquery.Selection) int {
	value, exists := node.Attr("height")
	if exists {
		nvalue, _ := strconv.Atoi(value)
		return nvalue
	}
	return 0
}

func (ve *VideoExtractor) getSrc(node *goquery.Selection) string {
	value, exists := node.Attr("src")
	if exists {
		return value
	}
	return ""
}

func (ve *VideoExtractor) getProvider(src string) string {
	if src != "" {
		for _, provider := range videoProviders {
			if strings.Contains(src, provider) {
				return provider
			}
		}
	}
	return ""
}

func (ve *VideoExtractor) getVideo(node *goquery.Selection) video {
	src := ve.getSrc(node)
	video := video{
		embedCode: ve.getEmbedCode(node),
		embedType: node.Get(0).DataAtom.String(),
		width:     ve.getWidth(node),
		height:    ve.getHeight(node),
		src:       src,
		provider:  ve.getProvider(src),
	}
	return video
}

func (ve *VideoExtractor) getIFrame(node *goquery.Selection) video {
	return ve.getVideo(node)
}

func (ve *VideoExtractor) getVideoTag(node *goquery.Selection) video {
	return video{}
}

func (ve *VideoExtractor) getEmbedTag(node *goquery.Selection) video {
	parent := node.Parent()
	if parent != nil {
		parentTag := parent.Get(0).DataAtom.String()
		if parentTag == "object" {
			return ve.getObjectTag(node)
		}
	}
	return ve.getVideo(node)
}

func (ve *VideoExtractor) getObjectTag(node *goquery.Selection) video {
	childEmbedTag := node.Find("embed")
	if ve.candidates.Has(childEmbedTag) {
		ve.candidates.Remove(childEmbedTag)
	}
	srcNode := node.Find(`param[name="movie"]`)
	if srcNode == nil || srcNode.Length() == 0 {
		return video{}
	}

	src, _ := srcNode.Attr("value")
	provider := ve.getProvider(src)
	if provider == "" {
		return video{}
	}
	video := ve.getVideo(node)
	video.provider = provider
	video.src = src
	return video
}

func (ve *VideoExtractor) GetVideos(article *Article) *set.Set {
	doc := article.Doc
	var nodes *goquery.Selection
	for _, videoTag := range videoTags {
		tmpNodes := doc.Find(videoTag)
		if nodes == nil {
			nodes = tmpNodes
		} else {
			nodes.Union(tmpNodes)
		}
	}

	nodes.Each(func(i int, node *goquery.Selection) {
		tag := node.Get(0).DataAtom.String()
		var movie video
		switch tag {
		case "video":
			movie = ve.getVideoTag(node)
			break
		case "embed":
			movie = ve.getEmbedTag(node)
			break
		case "object":
			movie = ve.getObjectTag(node)
			break
		case "iframe":
			movie = ve.getIFrame(node)
			break
		default:
			{
			}
		}

		if movie.src != "" {
			ve.movies.Add(movie)
		}
	})

	return ve.movies
}
