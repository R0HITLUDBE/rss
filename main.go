package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/R0HITLUDBE/rssagg/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct{
	DB *database.Queries
}

func main() {

	godotenv.Load(".env")
	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT is not found in the envirornment")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB URL is not found in the envirornment")
	}

	conn, err := sql.Open("postgres",dbURL)
	if err != nil{
		log.Fatal("Cant connect to database", err)
	}

	db := database.New(conn)

	apiCfg := apiConfig{
		DB: db,
	}


	go startScraping(db, 10, time.Hour)



	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
    // AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
    AllowedOrigins:   []string{"https://*", "http://*"},
    // AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
    ExposedHeaders:   []string{"Link"},
    AllowCredentials: false,
    MaxAge:           300, // Maximum value not ignored by any of major browsers
  }))

		v1Router := chi.NewRouter()

		v1Router.Get("/health", handlerReadiness)
		v1Router.Get("/err",handleErr)
		v1Router.Post("/users",apiCfg.handlerCreateUser)
		v1Router.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerGetUser))
		v1Router.Post("/feeds", apiCfg.middlewareAuth(apiCfg.handlerCreateFeed))
		v1Router.Get("/feeds", apiCfg.handlerGetFeeds)
		v1Router.Post("/feed_follows",apiCfg.middlewareAuth(apiCfg.handlerCreateFeedFollow))
		v1Router.Get("/feed_follows",apiCfg.middlewareAuth(apiCfg.handlerGetFeedFollows))
		v1Router.Delete("/feed_follows/{feedFollowID}",apiCfg.middlewareAuth(apiCfg.handlerDeleteFeedFollow))
		v1Router.Get("/subscribed_posts",apiCfg.middlewareAuth(apiCfg.handlerGetPostForUser))
		v1Router.Get("/posts",apiCfg.handlerGetPosts)
		router.Mount("/v1", v1Router)

	srv := &http.Server{
		Handler: router,
		Addr: ":" + portString,
	}
	log.Printf("Server started on port %v", portString)

	err = srv.ListenAndServe()
	if err != nil{
		log.Fatal(err)
	}
	fmt.Println("PoRt:", portString)
}
