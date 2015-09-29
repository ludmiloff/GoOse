package goose

type Goose struct {
	Config Configuration
}

func New(args ...string) Goose {

	return Goose{
		Config: GetDefualtConfiguration(args...),
	}
}

func (this Goose) ExtractFromUrl(url string) (*Article, error) {
	cc := NewCrawler(this.Config, url, "")
	return cc.Crawl()
}

func (this Goose) ExtractFromRawHtml(url string, rawHtml string) (*Article, error) {
	cc := NewCrawler(this.Config, url, rawHtml)
	return cc.Crawl()
}
