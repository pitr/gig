package main

import (
	"github.com/pitr/gig"
	"github.com/pitr/gig/middleware"
)

func main() {
	g := gig.New()

	g.Use(middleware.Logger())

	g.Static("", "astrobotany/")

	plant := g.Group("/plant", middleware.CertAuth(middleware.ValidateHasCertificate))
	{
		plant.Handle("", func(c gig.Context) error {
			return c.Gemini(gig.StatusSuccess, "Hello "+c.Get("subject").(string))
		})
		plant.Handle("/water", func(c gig.Context) error {
			return c.NoContent(gig.StatusRedirectTemporary, "/plant")
		})
		plant.Handle("/name", func(c gig.Context) error {
			if name := c.QueryString(); name != "" {
				return c.NoContent(gig.StatusRedirectTemporary, "/plant")
			}
			return c.NoContent(gig.StatusInput, "Enter a new nickname for your plant")
		})
	}

	panic(g.Run(":1965", "astro.crt", "astro.key"))
}
