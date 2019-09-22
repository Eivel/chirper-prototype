package main

import (
	"chirperweb/internal/chirper"
	"chirperweb/internal/routing"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
)

const role = "admin"

func init() {
	godotenv.Load()
}

func main() {
	provider, err := chirper.NewDefaultPostgresProvider()
	provider.EnsureSchema()

	if err != nil {
		log.Fatal(err)
	}

	adminRouter := chi.NewRouter()
	adminRouter.Use(routing.CheckPermissions(role))
	adminRouter.Get("/chirps/count", chirper.CountChirpsHandler(provider))

	userRouter := chi.NewRouter()
	userRouter.Get("/chirps", chirper.GetChirpsHandler(provider))
	userRouter.Post("/chirps", chirper.CreateChirpHandler(provider))

	baseRouter := routing.NewBaseRouter()

	baseRouter.Mount("/v1/api/admin", adminRouter)
	baseRouter.Mount("/v1/api", userRouter)

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("%s %s\n", method, route)
		return nil
	}
	if err := chi.Walk(baseRouter, walkFunc); err != nil {
		log.Panicf("Logging err: %s\n", err.Error())
	}

	log.Fatal(http.ListenAndServe(":8080", baseRouter))
}
