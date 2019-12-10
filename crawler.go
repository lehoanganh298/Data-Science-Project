package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/gocolly/colly"
)

func main() {
	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.imdb.com", "imdb.com"),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  ".*imdb.*",
		Parallelism: 5,
		Delay:       10 * time.Second,
	})

	// Create more collector
	movieCollector := c.Clone()
	userCollector := c.Clone()
	movieCount := 0
	userCount := 0

	userSet := mapset.NewSet()
	userSet.Add("abc")

	// Collect movieID and past for userCollector
	movieCollector.OnHTML(".lister-list tr:nth-child(-n+5)", func(e *colly.HTMLElement) {
		href := e.ChildAttr("a", "href")
		titleID := regexp.MustCompile(`/title/(\w+)/.*`).FindStringSubmatch(href)[1]
		userCollector.Visit(fmt.Sprintf("https://www.imdb.com/title/%s/reviews", titleID))
	})

	// Collect userID and past for rattingCollector
	userCollector.OnHTML(".lister-item-content .display-name-link a[href]", func(e *colly.HTMLElement) {
		userCount++
	})
	loadMore := 0
	userCollector.OnHTML("div[data-key]", func(e *colly.HTMLElement) {
		dataKey := e.Attr("data-key")
		rURL := e.Request.URL
		if !strings.Contains(rURL.Path, "_ajax") {
			rURL.Path = rURL.Path + "/_ajax"
		}
		q := rURL.Query()
		fmt.Println("load more")
		loadMore++
		q.Set("ref_", "undefined")
		q.Set("paginationKey", dataKey)
		rURL.RawQuery = q.Encode()
		userCollector.Visit(rURL.String())
	})

	// Before making a request print "Visiting ..."
	movieCollector.OnRequest(func(r *colly.Request) {
		fmt.Println("MovieCollector visiting:", r.URL.String())
	})
	userCollector.OnRequest(func(r *colly.Request) {
		// fmt.Println("UserCollector visiting:", r.URL.String())
	})

	movieCollector.Visit("https://www.imdb.com/chart/top")
	movieCollector.Wait()
	userCollector.Wait()
	fmt.Println(movieCount)
	fmt.Println(userCount)
	fmt.Println(loadMore)
}
