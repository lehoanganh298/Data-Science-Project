package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/gocolly/colly"
)

func main() {

	// Only process users in certain range to avoid request limit
	beginIndex := 100
	endIndex := 200

	//***********************************
	frating, err := os.Create(fmt.Sprintf("rating%d-%d.csv", beginIndex, endIndex))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer frating.Close()
	frating.WriteString("UserID, MovieID, Rating \n")

	//********************************************************
	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.imdb.com", "imdb.com"),
		colly.Async(true),
	)
	c.Limit(&colly.LimitRule{
		DomainGlob:  ".*imdb.*",
		Parallelism: 10,
		Delay:       1 * time.Second,
	})

	// Create more collector
	ratingCollector := c.Clone()
	ratingCount := 0
	userCount := 0
	doneUserCount := 0 //Count number of user reach the last review page (checking for completeness)

	//*************************************************************************
	// Visit each user's page and collect the ratings
	ratingCollector.OnRequest(func(r *colly.Request) {
		// fmt.Println("ratingCollector visiting:", r.URL.String())
		fmt.Print(".")
	})
	ratingCollector.OnHTML(".lister-item", func(e *colly.HTMLElement) {
		// Get userID
		re := regexp.MustCompile(`https://www.imdb.com/user/(\w+)/ratings`).FindStringSubmatch(e.Request.URL.String())
		if len(re) < 2 {
			fmt.Println("cannot get userID in this link: ", e.Request.URL.String())
			return
		}
		userID := re[1]

		// Get movieID
		titleHref := e.ChildAttr(".lister-item-header a", "href")
		re = regexp.MustCompile(`/title/(\w+)/.*`).FindStringSubmatch(titleHref)
		if len(re) < 2 {
			fmt.Println("\ncannot get titleID in this link: ", titleHref)
			return
		}
		movieID := re[1]

		// Get rating
		ratingStr := e.ChildText(".ipl-rating-star--other-user .ipl-rating-star__rating")
		rating, err := strconv.Atoi(ratingStr)
		if err != nil {
			fmt.Println("\ncannot get rating in this item:", userID, " ", movieID, ": ", ratingStr)
			fmt.Println(e.Request.URL.String())
			return
		}

		// Write to file
		frating.WriteString(fmt.Sprintf("%s,%s,%d\n", userID, movieID, rating))
		ratingCount++
	})
	ratingCollector.OnHTML("a.next-page", func(e *colly.HTMLElement) {
		// fmt.Println(e.ChildText(".pagination-range"))
		linkNextPage := e.Attr("href")
		if linkNextPage == "#" {
			doneUserCount++
		} else {
			ratingCollector.Visit("https://www.imdb.com" + linkNextPage)
		}
	})
	ratingCollector.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
		time.Sleep(10 * time.Second)
		ratingCollector.Visit(r.Request.URL.String())
	})

	//************************************
	// Load user list
	fuser, err := os.Open("..\\crawl-users\\users.txt")
	if err != nil {
		fmt.Println(err)
	}
	defer fuser.Close()

	userIndex := 0
	scanner := bufio.NewScanner(fuser)
	for scanner.Scan() {
		userID := scanner.Text()
		correctUserID, _ := regexp.MatchString(`ur[0-9]+`, userID)
		if correctUserID {
			if userIndex >= beginIndex && userIndex < endIndex {
				userCount++
				ratingCollector.Visit(fmt.Sprintf("https://www.imdb.com/user/%s/ratings", userID))
				fmt.Println("User ", userIndex, ": ", userID)
			}
			userIndex++
		}
	}

	ratingCollector.Wait()
	fmt.Println("\nNum users: ", userCount)
	fmt.Println("Num ratings: ", ratingCount)
	fmt.Println("Num done users:", doneUserCount)
}
