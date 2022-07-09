package main

import (
	"os"

	"github.com/pitr/gig"
)

func main() {
	g := gig.Default()
	g.ReadTimeout = time.Second * 10
	g.AllowProxying = true
	g.Use(gig.Titan(1024))

	g.Handle("/file.txt*", func(c gig.Context) error {
		if c.Get("titan").(bool) {
			data, err := gig.TitanReadFull(c)
			if err != nil {
				return err
			}

			err = os.WriteFile("file.txt", data, 0o644)
			if err != nil {
				return err
			}
			gig.TitanRedirect(c)
			return nil
		}
		return c.File("file.txt")
	})

	g.Run("../astro.crt", "../astro.key")
}
