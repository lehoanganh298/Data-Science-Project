package main

import (
	"fmt"
	"strings"
	"time"

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

	// Collect movieID and past for userCollector
	movieCollector.OnHTML(".titleColumn", func(e *colly.HTMLElement) {
		href := e.ChildAttr("a", "href")
		movieCount++
		titleID := strings.Split(href, "/")[2]
		userCollector.Visit(fmt.Sprintf("https://www.imdb.com/title/%s/reviews", titleID))
	})

	// Collect userID and past for rattingCollector
	userCollector.OnHTML(".lister-item-content", func(e *colly.HTMLElement) {
		userCount++
	})
	userCollector.OnHTML("div[data-key]", func(e *colly.HTMLElement) {
		dataKey := e.Attr("data-key")
		rURL := e.Request.URL
		if !strings.Contains(rURL.Path, "_ajax") {
			rURL.Path = rURL.Path + "/_ajax"
		}
		q := rURL.Query()
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
		fmt.Println("UserCollector visiting:", r.URL.String())
	})

	movieCollector.Visit("https://www.imdb.com/chart/top")
	movieCollector.Wait()
	userCollector.Wait()
	fmt.Println(movieCount)
	fmt.Println(userCount)
}
