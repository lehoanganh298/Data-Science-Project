package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

func main() {
	// Get from https://www.imdb.com/feature/genre

	genre_list := [24]string{"Action", "Adventure", "Animation", "Biography", "Comedy", "Crime",
		"Documentary", "Drama", "Family", "Fantasy", "Film Noir", "History",
		"Horror", "Music", "Musical", "Mystery", "Romance", "Sci-Fi",
		"Short Film", "Sport", "Superhero", "Thriller", "War", "Western"}

	// genre_map := make(map[string]int)
	// for idx, genre := range genre_list {
	// 	genre_map[genre] = idx
	// }

	// var output_list [38720]string
	// output_idx := 0

	//***********************************
	beginIndex := 0
	endIndex := 40000
	// finfo, err := os.Create(fmt.Sprintf("movies-info%d-%d.csv", beginIndex, endIndex))
	finfo, err := os.Create("movies-left.csv")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer finfo.Close()

	finfo.WriteString("MovieID,")
	for _, genre := range genre_list {
		finfo.WriteString(fmt.Sprintf("%s,", genre))
	}
	finfo.WriteString("\n")

	// Instantiate default collector
	movieInfoCollector := colly.NewCollector(
		colly.AllowedDomains("www.imdb.com", "imdb.com"),
		colly.Async(true),
		colly.AllowURLRevisit(),
	)
	movieInfoCollector.Limit(&colly.LimitRule{
		DomainGlob:  ".*imdb.*",
		Parallelism: 10,
	})
	movieInfoCollector.SetRequestTimeout(60 * time.Second)

	//*****************************************************

	movieInfoCollector.OnRequest(func(r *colly.Request) {
		fmt.Print(".")
	})
	movieInfoCollector.OnHTML("#titleStoryLine div.see-more", func(e *colly.HTMLElement) {
		text := strings.TrimSpace(e.Text)
		if text[:7] != "Genres:" {
			return
		}

		re := regexp.MustCompile(`https://www.imdb.com/title/(\w+)/`).FindStringSubmatch(e.Request.URL.String())
		if len(re) < 2 {
			fmt.Println("cannot get userID in this link: ", e.Request.URL.String())
			return
		}
		movieID := re[1]

		// fmt.Println(movieID)
		split1 := strings.Split(text, ":")
		if len(split1) != 2 {
			fmt.Println(movieID, "---", text, "***")
			return
		}

		genres := strings.Split(split1[1], "|")
		genre_map := make(map[string]int)
		for _, genre := range genres {
			genre_map[strings.TrimSpace(genre)] = 1
		}

		s := fmt.Sprintf("%s,", movieID)

		for _, genre := range genre_list {
			s += fmt.Sprintf("%d,", genre_map[genre])
		}
		s += "\n"
		finfo.WriteString(s)
		// fmt.Println("*************888")
		// fmt.Println(s)
		// output_list[output_idx] = s
		// output_idx++
	})
	movieInfoCollector.OnError(func(r *colly.Response, err error) {
		// fmt.Println("Request URL:", r.Request.URL, "\nError:", err.Error())
		fmt.Print("*", r.StatusCode)
		if r.StatusCode == 503 {
			time.Sleep(3 * time.Minute)
		} else {
			time.Sleep(30 * time.Second)
		}
		movieInfoCollector.Visit(r.Request.URL.String())
	})

	//************************************
	// Load movie list
	f, err := os.Open("movie-left.txt")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	idx := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		idx++

		movieID := scanner.Text()
		fmt.Println(movieID)
		correctmovieID, _ := regexp.MatchString(`tt[0-9]+`, movieID)
		if correctmovieID {
			if idx >= beginIndex && idx < endIndex {
				if idx%300 == 0 {
					movieInfoCollector.Wait()
					fmt.Println(idx)
					time.Sleep(1 * time.Second)
				}
				movieInfoCollector.Visit(fmt.Sprintf("https://www.imdb.com/title/%s/", movieID))
				// fmt.Println(movieID)
			}
		}

	}

	movieInfoCollector.Wait()
	// for _, output := range output_list {
	// 	// if i > 100 {
	// 	// 	break
	// 	// }
	// 	// fmt.Println("*", output, "*")
	// 	finfo.WriteString(output)
	// }
}
