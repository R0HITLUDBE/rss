package main

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/R0HITLUDBE/rssagg/internal/database"
	"github.com/google/uuid"
)

func startScraping(
	db *database.Queries,
	concurrency int,
	timeBetweenRequest time.Duration,
){
	log.Printf("Scraping on %v goroutines every %s duration",concurrency, timeBetweenRequest)

	ticker := 	time.NewTicker(timeBetweenRequest)

	for ; ; <-ticker.C{
		feeds, err :=	db.GetNextFeedsToFetch(context.Background(),int32(concurrency))

		if err != nil{
			log.Println("error fetching feed: ",feeds)
			continue
		}

		wg := &sync.WaitGroup{}
		for _, feed := range feeds{
			wg.Add(1)

			go scrapeFeed(db, wg, feed)
		}
		wg.Wait()
	}
}

// adding one for every feed , if we had concurrency of 30 we will be adding 30 to wait group and 30 feeds will be fetched concurrently

func scrapeFeed(db *database.Queries ,wg *sync.WaitGroup, feed database.Feed){

	defer wg.Done()

	_ , err := db.MarkFeedAsFetched(context.Background(), feed.ID)
	if err != nil{
		log.Println("Error making feed as fetched: ", err)
	}

	rssFeed, err := 	urlToFeed(feed.Url)
	if err != nil{
		log.Println("Error fetching feed : ", err)
		return
	}

	for _, item := range rssFeed.Channel.Item{
		description := sql.NullString{}
		if item.Description != ""{
			description.String = item.Description
			description.Valid = true
		}

		_, err := db.CreatePost(context.Background(),
			database.CreatePostParams{
			ID: uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Title: item.Title,
			Description: description,
			PublishedAt: item.PubDate,
			Url: item.Link,
			FeedID: feed.ID,
		})
		if err != nil{
			if strings.Contains(err.Error(),"duplicate key" ){
				continue
			}
			log.Println("Error fetching feed: ", err)
		}
	}
	log.Printf("Feed %s collected, %v post found", feed.Name, len(rssFeed.Channel.Item))

}