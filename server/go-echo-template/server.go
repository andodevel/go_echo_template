package server

import (
	"log"
	"net/http"
	"strings"

	"github.com/andodevel/go-echo-template/helpers"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/andodevel/go-echo-template/bootstrap"
	"github.com/andodevel/go-echo-template/constants"
	"github.com/andodevel/go-echo-template/db"
	"github.com/andodevel/go-echo-template/graphql"
	"github.com/andodevel/go-echo-template/server/routes"
)

// Start ...
func Start() {
	log.Println("Start with profile " + strings.ToUpper(bootstrap.GetProfile()))
	if bootstrap.IsDebugEnabled() {
		log.Println("Debug mode was enabled!")
	} else {
		log.Println("Release mode was enabled!")
	}

	db := db.CurrentDBConn()

	// Create new echo instance
	e := echo.New()

	// TODO: Move middlewares logic to middlewares package and group
	// Middlwares
	if bootstrap.IsDebugEnabled() {
		e.Use(middleware.Logger())
	}
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))
	// TODO: Add JWT middleware?
	// TODO: Move serect key to config
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))

	// Routes
	// TODO: Replace static templates with Reactjs
	e.GET("/", func(c echo.Context) error {
		tokenCookie, _ := c.Cookie(helpers.JWTCookieKey)

		vars := helpers.Map{"username": "Guest User"}
		if tokenCookie != nil {
			claims, _ := helpers.ParseJWTToken(tokenCookie.Value)
			if claims != nil {
				vars["username"] = claims.Username
			}
		}
		var html, _ = helpers.ParseHTMLTemplateFile("index", "server/go-echo-template/templates/index.html", vars)
		return c.HTML(http.StatusOK, html)
	})

	if bootstrap.IsInDevMode() {
		e.GET("/gql", playgroundHandler())
	}
	e.POST("/gql/query", graphqlHandler(db))

	routes.AuthGroup(e)

	// Start
	envPort := bootstrap.Prop(constants.EnvPort)
	if "" == envPort {
		envPort = "8080"
	}
	e.Logger.Fatal(e.Start(":" + envPort))
}

func playgroundHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		h := handler.Playground("GraphQL Playground", "/gql/query")
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}

func graphqlHandler(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		h := handler.GraphQL(graphql.NewExecutableSchema(graphql.NewGormConfig(db)))
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}
