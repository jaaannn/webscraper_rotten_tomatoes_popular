package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/gocolly/colly"
)

type RT_Show struct {
	title, avg_audience_score, avg_critic_score, rt_url, synopsis, genre string
}

func main() {
	var wg sync.WaitGroup
	c := colly.NewCollector()

	c.OnHTML("div.flex-container", func(e *colly.HTMLElement) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			url := "https://www.rottentomatoes.com" + e.ChildAttr("a", "href")
			c.Visit(url)
		}()
	})

	var rt_shows []RT_Show

	c.OnHTML("div.container.rt-layout__body", func(e *colly.HTMLElement) {
		rt_show := RT_Show{}

		rt_show.title = e.ChildText("h1.unset[slot=titleIntro]")
		rt_show.rt_url = e.Request.URL.String()
		rt_show.avg_audience_score = e.ChildText("rt-button[slot=audienceScore]")
		rt_show.avg_critic_score = e.ChildText("rt-button[slot=criticsScore]")
		rt_show.synopsis = e.ChildText("div.synopsis-wrap rt-text:not(.key)")

		e.ForEach("div.category-wrap", func(_ int, h *colly.HTMLElement) {
			if h.ChildText("dt") == "Genre" {
				rt_show.genre = h.ChildText("rt-link")
			}
		})

		rt_shows = append(rt_shows, rt_show)
	})

	c.OnError(func(r *colly.Response, e error) {
		fmt.Println("An error occurred!:", e)
	})

	c.Visit("https://www.rottentomatoes.com/browse/tv_series_browse/sort:popular")
	wg.Wait()

	if err := write_RT_Shows_to_CSV(rt_shows); err != nil {
		log.Fatalf("Failed to write to CSV: %v", err)
	}
}

func write_RT_Shows_to_CSV(shows []RT_Show) error {
	file, err := os.Create("rt_shows.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{
		"title",
		"genre",
		"avg_audience_score",
		"avg_tomatometer",
		"synopsis",
		"rt_url",
	}

	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, rt_show := range shows {
		if rt_show.title == "" {
			continue
		}
		record := []string{
			rt_show.title,
			rt_show.genre,
			rt_show.avg_audience_score,
			rt_show.avg_critic_score,
			rt_show.synopsis,
			rt_show.rt_url,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}
