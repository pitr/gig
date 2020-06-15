package main

import "github.com/pitr/gig"

func main() {
	g := gig.Default()

	g.Static("", "astrobotany/")

	plant := g.Group("/plant", gig.CertAuth(gig.ValidateHasCertificate))
	{
		plant.Handle("", func(c gig.Context) error {
			return c.Gemini("Hello " + c.Get("subject").(string))
		})
		plant.Handle("/water", func(c gig.Context) error {
			return c.NoContent(gig.StatusRedirectTemporary, "/plant")
		})
		plant.Handle("/name", func(c gig.Context) error {
			if name, err := c.QueryString(); err != nil {
				return c.NoContent(gig.StatusInput, "Bad input, try again")
			} else if name != "" {
				return c.NoContent(gig.StatusRedirectTemporary, "/plant")
			}
			return c.NoContent(gig.StatusInput, "Enter a new nickname for your plant")
		})
	}

	panic(g.Run(":1965", "astro.crt", "astro.key"))
}
