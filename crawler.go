package goose

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strings"

	"golang.org/x/text/transform"
	"golang.org/x/text/encoding/charmap"
	"github.com/PuerkitoBio/goquery"
)

type Crawler struct {
	config Configuration
	url     string
	rawHtml string
	helper  Helper
}

func NewCrawler(config Configuration, url string, rawHtml string) Crawler {
	return Crawler{
		config:  config,
		url:     url,
		rawHtml: rawHtml,
	}
}

func (this Crawler) Crawl() (*Article, error) {

	article := new(Article)
	this.assignParseCandidate()
	this.assignHtml()

	if this.rawHtml == "" {
		return article, nil
	}

	reader := strings.NewReader(this.rawHtml)
	document, err := goquery.NewDocumentFromReader(reader)

	if err != nil {
		return article, err
	}

	attr := ""
	selection := document.Find("meta").EachWithBreak(func(i int, s *goquery.Selection) bool {
		attr, exists := s.Attr("http-equiv")
		if exists && attr == "Content-Type" {
			return false
		}
		return true
	})

	if selection != nil {
		attr, _ = selection.Attr("content")
		attr = strings.Replace(attr, " ", "", -1)

		if strings.HasPrefix(attr, "text/html;charset=") {
			cs := strings.TrimPrefix(attr, "text/html;charset=")
			cs = strings.ToLower(cs)

			if cs != "utf-8" {
				
				sr := strings.NewReader(this.rawHtml)
				tr := transform.NewReader(sr, charmap.Windows1251.NewDecoder())
				utf8, err := ioutil.ReadAll(tr)
				if err != nil {
					// обработка на грешките
					this.rawHtml = ""
				} else {
					this.rawHtml = string(utf8)
				}				
				
				reader = strings.NewReader(this.rawHtml)
				document, err = goquery.NewDocumentFromReader(reader)
			}
		}
	}

	if err == nil {
		extractor := NewExtractor(this.config)
		html, _ := document.Html()
		start := TimeInNanoseconds()
		article.RawHtml = html
		article.FinalUrl = this.helper.url
		article.LinkHash = this.helper.linkHash
		article.Doc = document
		article.Title = extractor.getTitle(article)
		article.MetaLang = extractor.getMetaLanguage(article)
		article.MetaFavicon = extractor.getFavicon(article)

		article.MetaDescription = extractor.getMetaContentWithSelector(article, "meta[name#=(?i)^description$]")
		article.MetaKeywords = extractor.getMetaContentWithSelector(article, "meta[name#=(?i)^keywords$]")
		article.CanonicalLink = extractor.getCanonicalLink(article)
		article.Domain = extractor.getDomain(article)
		article.Tags = extractor.getTags(article)

		cleaner := NewCleaner(this.config)
		article.Doc = cleaner.clean(article)

		article.TopImage = OpenGraphResolver(article)
		if article.TopImage == "" {
			article.TopImage = WebPageResolver(article)
		}
		article.TopNode = extractor.calculateBestNode(article)
		if article.TopNode != nil {
			article.TopNode = extractor.postCleanup(article.TopNode)

			outputFormatter := new(outputFormatter)
			article.CleanedText = outputFormatter.getFormattedText(article)

			videoExtractor := NewVideoExtractor()
			article.Movies = videoExtractor.GetVideos(article)
		}

		stop := TimeInNanoseconds()
		delta := stop - start
		article.Delta = delta

	} else {
		return article, err
	}
	return article, err
}

func (this *Crawler) assignParseCandidate() {
	if this.rawHtml != "" {
		this.helper = NewRawHelper(this.url, this.rawHtml)
	} else {
		this.helper = NewUrlHelper(this.url)
	}
}

func (this *Crawler) assignHtml() {
	if this.rawHtml == "" {
		cookieJar, _ := cookiejar.New(nil)
		client := &http.Client{
			Jar:     cookieJar,
			Timeout: this.config.timeout,
		}
		req, err := http.NewRequest("GET", this.url, nil)
		if err == nil {
			req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_6_7) AppleWebKit/534.30 (KHTML, like Gecko) Chrome/12.0.742.91 Safari/534.30")
			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				contents, err := ioutil.ReadAll(resp.Body)
				if err == nil {
					this.rawHtml = string(contents)
				} else {
					log.Println(err.Error())
				}
			} else {
				log.Println(err.Error())
			}
		} else {
			log.Println(err.Error())
		}
	}
}
