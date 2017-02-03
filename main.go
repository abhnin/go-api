package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
	"twreporter.org/go-api/routers"
)

func main() {
	// security: no one can put it in an iframe
	secureMiddleware := secure.New(secure.Options{
		FrameDeny: true,
	})
	secureFunc := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			err := secureMiddleware.Process(c.Writer, c.Request)

			// If there was an error, do not continue.
			if err != nil {
				c.Abort()
				return
			}

			// Avoid header rewrite if response is a redirection.
			if status := c.Writer.Status(); status > 300 && status < 399 {
				c.Abort()
			}
		}
	}()

	router := routers.SetupRouter()

	router.Use(secureFunc)

	s := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}
	s.ListenAndServe()
}
