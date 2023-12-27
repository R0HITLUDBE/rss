package main

import (
	"fmt"
	"net/http"

	"github.com/R0HITLUDBE/rssagg/internal/auth"
	"github.com/R0HITLUDBE/rssagg/internal/database"
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (cfg *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := auth.GetApiKey(r.Header)
		if err != nil{
			respondWithError(w, 403, fmt.Sprintf("Auth error: %v",err))
			return
		}
		user, err := cfg.DB.GetUserByApiKey(r.Context(), apiKey)
		if err != nil{
			respondWithError(w, 400, fmt.Sprintf("Couldnt get user : %v",err))
			return
		}
		handler(w,r,user)
	}
}