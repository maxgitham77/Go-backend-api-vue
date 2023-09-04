package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/maxGitham77/vueapi/internal/data"
	"net/http"
	"time"
)

// routes generates our routes and attaches them to handlers, using the chi router
// Note that we return type http.Handler, and not *chi.Mux; since chi.Mux satisfies the
// interface requirements for http.Handler, it makes sense to return the type taht is part
// of the standard library
func (app *application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	mux.Post("/users/login", app.Login)
	mux.Post("/users/logout", app.Logout)

	mux.Get("/users/all", func(writer http.ResponseWriter, request *http.Request) {
		var users data.User
		all, err := users.GetAll()
		if err != nil {
			app.errorLog.Println(err)
			return
		}

		paylaod := jsonResponse{
			Error:   false,
			Message: "success",
			Data:    envelop{"users": all},
		}

		app.writeJSON(writer, http.StatusOK, paylaod)
	})

	mux.Get("/users/add", func(writer http.ResponseWriter, request *http.Request) {
		var u = data.User{
			Email:     "friend@friend.com",
			Firstname: "Friend",
			Lastname:  "Forever",
			Password:  "password",
		}

		app.infoLog.Println("Adding user...")

		id, err := app.models.User.Insert(u)
		if err != nil {
			app.errorLog.Println(err)
			app.errorJSON(writer, err, http.StatusForbidden)
			return
		}
		app.infoLog.Println("Got back id of ", id)
		newUser, _ := app.models.User.GetById(id)
		app.writeJSON(writer, http.StatusOK, newUser)
	})

	mux.Get("/test-generate-token", func(writer http.ResponseWriter, request *http.Request) {
		token, err := app.models.User.Token.GenerateToken(2, 60*time.Minute)
		if err != nil {
			app.errorLog.Println(err)
			return
		}

		token.Email = "friend@friend.com"
		token.CreatedAt = time.Now()
		token.UpdatedAt = time.Now()

		payload := jsonResponse{
			Error:   false,
			Message: "success",
			Data:    token,
		}
		app.writeJSON(writer, http.StatusOK, payload)

	})

	mux.Get("/test-save-token", func(writer http.ResponseWriter, request *http.Request) {
		token, err := app.models.User.Token.GenerateToken(2, 60*time.Minute)
		if err != nil {
			app.errorLog.Println(err)
			return
		}

		user, err := app.models.User.GetById(2)
		if err != nil {
			app.errorLog.Println(err)
			return
		}

		token.UserID = user.ID
		token.CreatedAt = time.Now()
		token.UpdatedAt = time.Now()

		err = token.Insert(*token, *user)
		if err != nil {
			app.errorLog.Println(err)
			return
		}

		payload := jsonResponse{
			Error:   false,
			Message: "success",
			Data:    token,
		}
		app.writeJSON(writer, http.StatusOK, payload)
	})

	mux.Get("/test-validate-token", func(writer http.ResponseWriter, request *http.Request) {
		tokenToValidate := request.URL.Query().Get("token")
		valid, err := app.models.Token.ValidToken(tokenToValidate)
		if err != nil {
			app.errorJSON(writer, err)
			return
		}
		var payload jsonResponse
		payload.Error = false
		payload.Data = valid

		app.writeJSON(writer, http.StatusOK, payload)
	})

	return mux
}
