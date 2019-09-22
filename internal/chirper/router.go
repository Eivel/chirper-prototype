package chirper

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/render"
)

// GetChirpsHandler handles the GET request for fetching chirps.
// Accepts the following parameters:
// tags - []string *optional.
func GetChirpsHandler(repository Repository) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tags := r.URL.Query()["tags"]
		chirps, err := repository.GetChirps(tags)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		render.JSON(w, r, chirps)
	})
}

// CreateChirpHandler handles the POST request for creating chirps.
// Accepts JSON body with the chirp of the following format:
// {"message": "", "tags": ["", ""], "author": ""}.
func CreateChirpHandler(repository Repository) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var chirp Chirp
		json.NewDecoder(r.Body).Decode(&chirp)

		err := repository.CreateChirp(chirp)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		response := make(map[string]string)
		response["message"] = "Chirp created successfully"
		render.JSON(w, r, response)
	})
}

// CountChirpsHandler handles the GET request for fetching chirps count.
// Accepts the following parameters:
// startingDate - string *required,
// endingDate - string *required,
// tags - []string *optional.
func CountChirpsHandler(repository Repository) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tags := r.URL.Query()["tags"]
		startingDate := r.URL.Query().Get("startingDate")
		endingDate := r.URL.Query().Get("endingDate")

		if startingDate == "" || endingDate == "" {
			http.Error(w, http.StatusText(400), 400)
			return
		}

		chirps, err := repository.CountChirps(startingDate, endingDate, tags)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		render.JSON(w, r, chirps)
	})
}
