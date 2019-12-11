package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/gocolly/colly"
)

const minReviewsPerUser = 100

func main() {
	//************************************
	// Load user list  if available
	f, err := os.OpenFile("users.txt",
		os.O_APPEND|os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	// Add user list to userSet
	userSet := mapset.NewSet()
	oldUserCount := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			userSet.Add(line)
			oldUserCount++
		}
	}
	fmt.Println("Number of users read from file: ", oldUserCount)

	//***********************************
	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.imdb.com", "imdb.com"),
		colly.Async(true),
	)
	c.Limit(&colly.LimitRule{
		DomainGlob:  ".*imdb.*",
		Parallelism: 10,
		Delay:       3 * time.Second,
	})

	// Create more collector
	movieCollector := c.Clone()
	userCollector := c.Clone()
	qualifiedUserCollector := c.Clone()
	movieCount := 0
	userCount := 0

	//**********************************************************************
	// Go through webpage with list of movies, collect movieID and past for userCollector
	movieCollector.OnRequest(func(r *colly.Request) {
		// fmt.Println("MovieCollector visiting:", r.URL.String())
	})
	movieCollector.OnHTML(".titleColumn", func(e *colly.HTMLElement) {
		href := e.ChildAttr("a", "href")
		movieCount++
		titleID := strings.Split(href, "/")[2]
		userCollector.Visit(fmt.Sprintf("https://www.imdb.com/title/%s/reviews", titleID))
	})
	movieCollector.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
		// time.Sleep(1 * time.Second)
		// userCollector.Visit(r.Request.URL.String())
	})

	//**********************************************************************
	// Visit each movie page, collect userIDs and pass to qualifiedUserCollector
	userCollector.OnRequest(func(r *colly.Request) {
		// fmt.Println(count, " UserCollector visiting:", r.URL.String())
	})
	userCollector.OnHTML(".lister-item-content .display-name-link", func(e *colly.HTMLElement) {
		// Get userID
		href := e.ChildAttr("a", "href")
		re := regexp.MustCompile(`/user/(\w+)/.*`).FindStringSubmatch(href)
		if len(re) < 2 {
			fmt.Println("cannot get user in this element")
			return
		}
		userID := re[1]

		if !userSet.Contains(userID) {
			qualifiedUserCollector.Visit(fmt.Sprintf("https://www.imdb.com/user/%s/ratings", userID))
		}
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
	userCollector.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
		// time.Sleep(1 * time.Second)
		// userCollector.Visit(r.Request.URL.String())
	})

	//**********************************************************************
	// Visit each user's page, check number of reviews by this user >= minReviewsPerUser
	qualifiedUserCollector.OnHTML("#lister-header-current-size", func(e *colly.HTMLElement) {
		reviewCountStr := strings.ReplaceAll(e.Text, ",", "")
		reviewCount, err := strconv.Atoi(reviewCountStr)
		if err != nil {
			fmt.Println(err)
			fmt.Println("cannot find number of review in URL: ", e.Request.URL)
			return
		}

		if reviewCount >= minReviewsPerUser {
			// Get userID
			re := regexp.MustCompile(`https://www.imdb.com/user/(\w+)/ratings`).FindStringSubmatch(e.Request.URL.String())
			if len(re) < 2 {
				fmt.Println("cannot get userID in this link: ", e.Request.URL.String())
				return
			}
			userID := re[1]
			fmt.Println(userID, "---", reviewCount)

			userSet.Add(userID)
			userCount++
			_, err := f.WriteString(fmt.Sprintln(userID))
			if err != nil {
				fmt.Println(err)
			}
		}
	})

	movieCollector.Visit("https://www.imdb.com/chart/top")
	movieCollector.Wait()
	userCollector.Wait()
	qualifiedUserCollector.Wait()
	fmt.Println("Movie number: ", movieCount)
	fmt.Println("Number of users added: ", userCount)
}
