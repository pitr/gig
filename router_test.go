package gig

import (
	"strings"
	"testing"

	"github.com/matryer/is"
)

var (
	staticRoutes = []*Route{
		{"/", ""},
		{"/cmd.html", ""},
		{"/code.html", ""},
		{"/contrib.html", ""},
		{"/contribute.html", ""},
		{"/debugging_with_gdb.html", ""},
		{"/docs.html", ""},
		{"/effective_go.html", ""},
		{"/files.log", ""},
		{"/gccgo_contribute.html", ""},
		{"/gccgo_install.html", ""},
		{"/go-logo-black.png", ""},
		{"/go-logo-blue.png", ""},
		{"/go-logo-white.png", ""},
		{"/go1.1.html", ""},
		{"/go1.2.html", ""},
		{"/go1.html", ""},
		{"/go1compat.html", ""},
		{"/go_faq.html", ""},
		{"/go_mem.html", ""},
		{"/go_spec.html", ""},
		{"/help.html", ""},
		{"/ie.css", ""},
		{"/install-source.html", ""},
		{"/install.html", ""},
		{"/logo-153x55.png", ""},
		{"/Makefile", ""},
		{"/root.html", ""},
		{"/share.png", ""},
		{"/sieve.gif", ""},
		{"/tos.html", ""},
		{"/articles/", ""},
		{"/articles/go_command.html", ""},
		{"/articles/index.html", ""},
		{"/articles/wiki/", ""},
		{"/articles/wiki/edit.html", ""},
		{"/articles/wiki/final-noclosure.go", ""},
		{"/articles/wiki/final-noerror.go", ""},
		{"/articles/wiki/final-parsetemplate.go", ""},
		{"/articles/wiki/final-template.go", ""},
		{"/articles/wiki/final.go", ""},
		{"/articles/wiki/get.go", ""},
		{"/articles/wiki/http-sample.go", ""},
		{"/articles/wiki/index.html", ""},
		{"/articles/wiki/Makefile", ""},
		{"/articles/wiki/notemplate.go", ""},
		{"/articles/wiki/part1-noerror.go", ""},
		{"/articles/wiki/part1.go", ""},
		{"/articles/wiki/part2.go", ""},
		{"/articles/wiki/part3-errorhandling.go", ""},
		{"/articles/wiki/part3.go", ""},
		{"/articles/wiki/test.bash", ""},
		{"/articles/wiki/test_edit.good", ""},
		{"/articles/wiki/test_Test.txt.good", ""},
		{"/articles/wiki/test_view.good", ""},
		{"/articles/wiki/view.html", ""},
		{"/codewalk/", ""},
		{"/codewalk/codewalk.css", ""},
		{"/codewalk/codewalk.js", ""},
		{"/codewalk/codewalk.xml", ""},
		{"/codewalk/functions.xml", ""},
		{"/codewalk/markov.go", ""},
		{"/codewalk/markov.xml", ""},
		{"/codewalk/pig.go", ""},
		{"/codewalk/popout.png", ""},
		{"/codewalk/run", ""},
		{"/codewalk/sharemem.xml", ""},
		{"/codewalk/urlpoll.go", ""},
		{"/devel/", ""},
		{"/devel/release.html", ""},
		{"/devel/weekly.html", ""},
		{"/gopher/", ""},
		{"/gopher/appenginegopher.jpg", ""},
		{"/gopher/appenginegophercolor.jpg", ""},
		{"/gopher/appenginelogo.gif", ""},
		{"/gopher/bumper.png", ""},
		{"/gopher/bumper192x108.png", ""},
		{"/gopher/bumper320x180.png", ""},
		{"/gopher/bumper480x270.png", ""},
		{"/gopher/bumper640x360.png", ""},
		{"/gopher/doc.png", ""},
		{"/gopher/frontpage.png", ""},
		{"/gopher/gopherbw.png", ""},
		{"/gopher/gophercolor.png", ""},
		{"/gopher/gophercolor16x16.png", ""},
		{"/gopher/help.png", ""},
		{"/gopher/pkg.png", ""},
		{"/gopher/project.png", ""},
		{"/gopher/ref.png", ""},
		{"/gopher/run.png", ""},
		{"/gopher/talks.png", ""},
		{"/gopher/pencil/", ""},
		{"/gopher/pencil/gopherhat.jpg", ""},
		{"/gopher/pencil/gopherhelmet.jpg", ""},
		{"/gopher/pencil/gophermega.jpg", ""},
		{"/gopher/pencil/gopherrunning.jpg", ""},
		{"/gopher/pencil/gopherswim.jpg", ""},
		{"/gopher/pencil/gopherswrench.jpg", ""},
		{"/play/", ""},
		{"/play/fib.go", ""},
		{"/play/hello.go", ""},
		{"/play/life.go", ""},
		{"/play/peano.go", ""},
		{"/play/pi.go", ""},
		{"/play/sieve.go", ""},
		{"/play/solitaire.go", ""},
		{"/play/tree.go", ""},
		{"/progs/", ""},
		{"/progs/cgo1.go", ""},
		{"/progs/cgo2.go", ""},
		{"/progs/cgo3.go", ""},
		{"/progs/cgo4.go", ""},
		{"/progs/defer.go", ""},
		{"/progs/defer.out", ""},
		{"/progs/defer2.go", ""},
		{"/progs/defer2.out", ""},
		{"/progs/eff_bytesize.go", ""},
		{"/progs/eff_bytesize.out", ""},
		{"/progs/eff_qr.go", ""},
		{"/progs/eff_sequence.go", ""},
		{"/progs/eff_sequence.out", ""},
		{"/progs/eff_unused1.go", ""},
		{"/progs/eff_unused2.go", ""},
		{"/progs/error.go", ""},
		{"/progs/error2.go", ""},
		{"/progs/error3.go", ""},
		{"/progs/error4.go", ""},
		{"/progs/go1.go", ""},
		{"/progs/gobs1.go", ""},
		{"/progs/gobs2.go", ""},
		{"/progs/image_draw.go", ""},
		{"/progs/image_package1.go", ""},
		{"/progs/image_package1.out", ""},
		{"/progs/image_package2.go", ""},
		{"/progs/image_package2.out", ""},
		{"/progs/image_package3.go", ""},
		{"/progs/image_package3.out", ""},
		{"/progs/image_package4.go", ""},
		{"/progs/image_package4.out", ""},
		{"/progs/image_package5.go", ""},
		{"/progs/image_package5.out", ""},
		{"/progs/image_package6.go", ""},
		{"/progs/image_package6.out", ""},
		{"/progs/interface.go", ""},
		{"/progs/interface2.go", ""},
		{"/progs/interface2.out", ""},
		{"/progs/json1.go", ""},
		{"/progs/json2.go", ""},
		{"/progs/json2.out", ""},
		{"/progs/json3.go", ""},
		{"/progs/json4.go", ""},
		{"/progs/json5.go", ""},
		{"/progs/run", ""},
		{"/progs/slices.go", ""},
		{"/progs/timeout1.go", ""},
		{"/progs/timeout2.go", ""},
		{"/progs/update.bash", ""},
	}

	gitHubAPI = []*Route{
		// OAuth Authorizations
		{"/authorizations", ""},
		{"/authorizations/:id", ""},
		{"/authorizations", ""},
		//{"/authorizations/clients/:client_id", ""},
		//{"/authorizations/:id", ""},
		{"/authorizations/:id", ""},
		{"/applications/:client_id/tokens/:access_token", ""},
		{"/applications/:client_id/tokens", ""},
		{"/applications/:client_id/tokens/:access_token", ""},

		// Activity
		{"/events", ""},
		{"/repos/:owner/:repo/events", ""},
		{"/networks/:owner/:repo/events", ""},
		{"/orgs/:org/events", ""},
		{"/users/:user/received_events", ""},
		{"/users/:user/received_events/public", ""},
		{"/users/:user/events", ""},
		{"/users/:user/events/public", ""},
		{"/users/:user/events/orgs/:org", ""},
		{"/feeds", ""},
		{"/notifications", ""},
		{"/repos/:owner/:repo/notifications", ""},
		{"/notifications", ""},
		{"/repos/:owner/:repo/notifications", ""},
		{"/notifications/threads/:id", ""},
		//{"/notifications/threads/:id", ""},
		{"/notifications/threads/:id/subscription", ""},
		{"/notifications/threads/:id/subscription", ""},
		{"/notifications/threads/:id/subscription", ""},
		{"/repos/:owner/:repo/stargazers", ""},
		{"/users/:user/starred", ""},
		{"/user/starred", ""},
		{"/user/starred/:owner/:repo", ""},
		{"/user/starred/:owner/:repo", ""},
		{"/user/starred/:owner/:repo", ""},
		{"/repos/:owner/:repo/subscribers", ""},
		{"/users/:user/subscriptions", ""},
		{"/user/subscriptions", ""},
		{"/repos/:owner/:repo/subscription", ""},
		{"/repos/:owner/:repo/subscription", ""},
		{"/repos/:owner/:repo/subscription", ""},
		{"/user/subscriptions/:owner/:repo", ""},
		{"/user/subscriptions/:owner/:repo", ""},
		{"/user/subscriptions/:owner/:repo", ""},

		// Gists
		{"/users/:user/gists", ""},
		{"/gists", ""},
		//{"/gists/public", ""},
		//{"/gists/starred", ""},
		{"/gists/:id", ""},
		{"/gists", ""},
		//{"/gists/:id", ""},
		{"/gists/:id/star", ""},
		{"/gists/:id/star", ""},
		{"/gists/:id/star", ""},
		{"/gists/:id/forks", ""},
		{"/gists/:id", ""},

		// Git Data
		{"/repos/:owner/:repo/git/blobs/:sha", ""},
		{"/repos/:owner/:repo/git/blobs", ""},
		{"/repos/:owner/:repo/git/commits/:sha", ""},
		{"/repos/:owner/:repo/git/commits", ""},
		//{"/repos/:owner/:repo/git/refs/*ref", ""},
		{"/repos/:owner/:repo/git/refs", ""},
		{"/repos/:owner/:repo/git/refs", ""},
		//{"/repos/:owner/:repo/git/refs/*ref", ""},
		//{"/repos/:owner/:repo/git/refs/*ref", ""},
		{"/repos/:owner/:repo/git/tags/:sha", ""},
		{"/repos/:owner/:repo/git/tags", ""},
		{"/repos/:owner/:repo/git/trees/:sha", ""},
		{"/repos/:owner/:repo/git/trees", ""},

		// Issues
		{"/issues", ""},
		{"/user/issues", ""},
		{"/orgs/:org/issues", ""},
		{"/repos/:owner/:repo/issues", ""},
		{"/repos/:owner/:repo/issues/:number", ""},
		{"/repos/:owner/:repo/issues", ""},
		//{"/repos/:owner/:repo/issues/:number", ""},
		{"/repos/:owner/:repo/assignees", ""},
		{"/repos/:owner/:repo/assignees/:assignee", ""},
		{"/repos/:owner/:repo/issues/:number/comments", ""},
		//{"/repos/:owner/:repo/issues/comments", ""},
		//{"/repos/:owner/:repo/issues/comments/:id", ""},
		{"/repos/:owner/:repo/issues/:number/comments", ""},
		//{"/repos/:owner/:repo/issues/comments/:id", ""},
		//{"/repos/:owner/:repo/issues/comments/:id", ""},
		{"/repos/:owner/:repo/issues/:number/events", ""},
		//{"/repos/:owner/:repo/issues/events", ""},
		//{"/repos/:owner/:repo/issues/events/:id", ""},
		{"/repos/:owner/:repo/labels", ""},
		{"/repos/:owner/:repo/labels/:name", ""},
		{"/repos/:owner/:repo/labels", ""},
		//{"/repos/:owner/:repo/labels/:name", ""},
		{"/repos/:owner/:repo/labels/:name", ""},
		{"/repos/:owner/:repo/issues/:number/labels", ""},
		{"/repos/:owner/:repo/issues/:number/labels", ""},
		{"/repos/:owner/:repo/issues/:number/labels/:name", ""},
		{"/repos/:owner/:repo/issues/:number/labels", ""},
		{"/repos/:owner/:repo/issues/:number/labels", ""},
		{"/repos/:owner/:repo/milestones/:number/labels", ""},
		{"/repos/:owner/:repo/milestones", ""},
		{"/repos/:owner/:repo/milestones/:number", ""},
		{"/repos/:owner/:repo/milestones", ""},
		//{"/repos/:owner/:repo/milestones/:number", ""},
		{"/repos/:owner/:repo/milestones/:number", ""},

		// Miscellaneous
		{"/emojis", ""},
		{"/gitignore/templates", ""},
		{"/gitignore/templates/:name", ""},
		{"/markdown", ""},
		{"/markdown/raw", ""},
		{"/meta", ""},
		{"/rate_limit", ""},

		// Organizations
		{"/users/:user/orgs", ""},
		{"/user/orgs", ""},
		{"/orgs/:org", ""},
		//{"/orgs/:org", ""},
		{"/orgs/:org/members", ""},
		{"/orgs/:org/members/:user", ""},
		{"/orgs/:org/members/:user", ""},
		{"/orgs/:org/public_members", ""},
		{"/orgs/:org/public_members/:user", ""},
		{"/orgs/:org/public_members/:user", ""},
		{"/orgs/:org/public_members/:user", ""},
		{"/orgs/:org/teams", ""},
		{"/teams/:id", ""},
		{"/orgs/:org/teams", ""},
		//{"/teams/:id", ""},
		{"/teams/:id", ""},
		{"/teams/:id/members", ""},
		{"/teams/:id/members/:user", ""},
		{"/teams/:id/members/:user", ""},
		{"/teams/:id/members/:user", ""},
		{"/teams/:id/repos", ""},
		{"/teams/:id/repos/:owner/:repo", ""},
		{"/teams/:id/repos/:owner/:repo", ""},
		{"/teams/:id/repos/:owner/:repo", ""},
		{"/user/teams", ""},

		// Pull Requests
		{"/repos/:owner/:repo/pulls", ""},
		{"/repos/:owner/:repo/pulls/:number", ""},
		{"/repos/:owner/:repo/pulls", ""},
		//{"/repos/:owner/:repo/pulls/:number", ""},
		{"/repos/:owner/:repo/pulls/:number/commits", ""},
		{"/repos/:owner/:repo/pulls/:number/files", ""},
		{"/repos/:owner/:repo/pulls/:number/merge", ""},
		{"/repos/:owner/:repo/pulls/:number/merge", ""},
		{"/repos/:owner/:repo/pulls/:number/comments", ""},
		//{"/repos/:owner/:repo/pulls/comments", ""},
		//{"/repos/:owner/:repo/pulls/comments/:number", ""},
		{"/repos/:owner/:repo/pulls/:number/comments", ""},
		//{"/repos/:owner/:repo/pulls/comments/:number", ""},
		//{"/repos/:owner/:repo/pulls/comments/:number", ""},

		// Repositories
		{"/user/repos", ""},
		{"/users/:user/repos", ""},
		{"/orgs/:org/repos", ""},
		{"/repositories", ""},
		{"/user/repos", ""},
		{"/orgs/:org/repos", ""},
		{"/repos/:owner/:repo", ""},
		//{"/repos/:owner/:repo", ""},
		{"/repos/:owner/:repo/contributors", ""},
		{"/repos/:owner/:repo/languages", ""},
		{"/repos/:owner/:repo/teams", ""},
		{"/repos/:owner/:repo/tags", ""},
		{"/repos/:owner/:repo/branches", ""},
		{"/repos/:owner/:repo/branches/:branch", ""},
		{"/repos/:owner/:repo", ""},
		{"/repos/:owner/:repo/collaborators", ""},
		{"/repos/:owner/:repo/collaborators/:user", ""},
		{"/repos/:owner/:repo/collaborators/:user", ""},
		{"/repos/:owner/:repo/collaborators/:user", ""},
		{"/repos/:owner/:repo/comments", ""},
		{"/repos/:owner/:repo/commits/:sha/comments", ""},
		{"/repos/:owner/:repo/commits/:sha/comments", ""},
		{"/repos/:owner/:repo/comments/:id", ""},
		//{"/repos/:owner/:repo/comments/:id", ""},
		{"/repos/:owner/:repo/comments/:id", ""},
		{"/repos/:owner/:repo/commits", ""},
		{"/repos/:owner/:repo/commits/:sha", ""},
		{"/repos/:owner/:repo/readme", ""},
		//{"/repos/:owner/:repo/contents/*path", ""},
		//{"/repos/:owner/:repo/contents/*path", ""},
		//{"/repos/:owner/:repo/contents/*path", ""},
		//{"/repos/:owner/:repo/:archive_format/:ref", ""},
		{"/repos/:owner/:repo/keys", ""},
		{"/repos/:owner/:repo/keys/:id", ""},
		{"/repos/:owner/:repo/keys", ""},
		//{"/repos/:owner/:repo/keys/:id", ""},
		{"/repos/:owner/:repo/keys/:id", ""},
		{"/repos/:owner/:repo/downloads", ""},
		{"/repos/:owner/:repo/downloads/:id", ""},
		{"/repos/:owner/:repo/downloads/:id", ""},
		{"/repos/:owner/:repo/forks", ""},
		{"/repos/:owner/:repo/forks", ""},
		{"/repos/:owner/:repo/hooks", ""},
		{"/repos/:owner/:repo/hooks/:id", ""},
		{"/repos/:owner/:repo/hooks", ""},
		//{"/repos/:owner/:repo/hooks/:id", ""},
		{"/repos/:owner/:repo/hooks/:id/tests", ""},
		{"/repos/:owner/:repo/hooks/:id", ""},
		{"/repos/:owner/:repo/merges", ""},
		{"/repos/:owner/:repo/releases", ""},
		{"/repos/:owner/:repo/releases/:id", ""},
		{"/repos/:owner/:repo/releases", ""},
		//{"/repos/:owner/:repo/releases/:id", ""},
		{"/repos/:owner/:repo/releases/:id", ""},
		{"/repos/:owner/:repo/releases/:id/assets", ""},
		{"/repos/:owner/:repo/stats/contributors", ""},
		{"/repos/:owner/:repo/stats/commit_activity", ""},
		{"/repos/:owner/:repo/stats/code_frequency", ""},
		{"/repos/:owner/:repo/stats/participation", ""},
		{"/repos/:owner/:repo/stats/punch_card", ""},
		{"/repos/:owner/:repo/statuses/:ref", ""},
		{"/repos/:owner/:repo/statuses/:ref", ""},

		// Search
		{"/search/repositories", ""},
		{"/search/code", ""},
		{"/search/issues", ""},
		{"/search/users", ""},
		{"/legacy/issues/search/:owner/:repository/:state/:keyword", ""},
		{"/legacy/repos/search/:keyword", ""},
		{"/legacy/user/search/:keyword", ""},
		{"/legacy/user/email/:email", ""},

		// Users
		{"/users/:user", ""},
		{"/user", ""},
		//{"/user", ""},
		{"/users", ""},
		{"/user/emails", ""},
		{"/user/emails", ""},
		{"/user/emails", ""},
		{"/users/:user/followers", ""},
		{"/user/followers", ""},
		{"/users/:user/following", ""},
		{"/user/following", ""},
		{"/user/following/:user", ""},
		{"/users/:user/following/:target_user", ""},
		{"/user/following/:user", ""},
		{"/user/following/:user", ""},
		{"/users/:user/keys", ""},
		{"/user/keys", ""},
		{"/user/keys/:id", ""},
		{"/user/keys", ""},
		//{"/user/keys/:id", ""},
		{"/user/keys/:id", ""},
	}

	parseAPI = []*Route{
		// Objects
		{"/1/classes/:className", ""},
		{"/1/classes/:className/:objectId", ""},
		{"/1/classes/:className/:objectId", ""},
		{"/1/classes/:className", ""},
		{"/1/classes/:className/:objectId", ""},

		// Users
		{"/1/users", ""},
		{"/1/login", ""},
		{"/1/users/:objectId", ""},
		{"/1/users/:objectId", ""},
		{"/1/users", ""},
		{"/1/users/:objectId", ""},
		{"/1/requestPasswordReset", ""},

		// Roles
		{"/1/roles", ""},
		{"/1/roles/:objectId", ""},
		{"/1/roles/:objectId", ""},
		{"/1/roles", ""},
		{"/1/roles/:objectId", ""},

		// Files
		{"/1/files/:fileName", ""},

		// Analytics
		{"/1/events/:eventName", ""},

		// Push Notifications
		{"/1/push", ""},

		// Installations
		{"/1/installations", ""},
		{"/1/installations/:objectId", ""},
		{"/1/installations/:objectId", ""},
		{"/1/installations", ""},
		{"/1/installations/:objectId", ""},

		// Cloud Functions
		{"/1/functions", ""},
	}

	googlePlusAPI = []*Route{
		// People
		{"/people/:userId", ""},
		{"/people", ""},
		{"/activities/:activityId/people/:collection", ""},
		{"/people/:userId/people/:collection", ""},
		{"/people/:userId/openIdConnect", ""},

		// Activities
		{"/people/:userId/activities/:collection", ""},
		{"/activities/:activityId", ""},
		{"/activities", ""},

		// Comments
		{"/activities/:activityId/comments", ""},
		{"/comments/:commentId", ""},

		// Moments
		{"/people/:userId/moments/:collection", ""},
		{"/people/:userId/moments/:collection", ""},
		{"/moments/:id", ""},
	}

	// handlerHelper created a function that will set a context key for assertion
	handlerHelper = func(key string, value int) func(c Context) error {
		return func(c Context) error {
			c.Set(key, value)
			c.Set("path", c.Path())
			return nil
		}
	}
)

func TestRouterEmpty(t *testing.T) {
	is := is.New(t)

	g := New()
	r := g.router
	path := ""
	r.Add(path, func(c Context) error {
		c.Set("path", path)
		return nil
	})
	c := g.NewContext(nil, nil, "", nil).(*context)
	r.Find(path, c)

	is.NoErr(c.handler(c))
	is.Equal(path, c.Get("path"))
}

func TestRouterStatic(t *testing.T) {
	g := New()
	r := g.router
	path := "/folders/a/files/gig.gif"
	r.Add(path, func(c Context) error {
		c.Set("path", path)
		return nil
	})
	c := g.NewContext(nil, nil, "", nil).(*context)
	r.Find(path, c)

	is := is.New(t)
	is.NoErr(c.handler(c))
	is.Equal(path, c.Get("path"))
}

func TestRouterParam(t *testing.T) {
	g := New()
	r := g.router
	r.Add("/users/:id", func(c Context) error {
		return nil
	})
	c := g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/1", c)

	is := is.New(t)
	is.Equal("1", c.Param("id"))
}

func TestRouterTwoParam(t *testing.T) {
	g := New()
	r := g.router
	r.Add("/users/:uid/files/:fid", func(Context) error {
		return nil
	})
	c := g.NewContext(nil, nil, "", nil).(*context)

	r.Find("/users/1/files/1", c)

	is := is.New(t)
	is.Equal("1", c.Param("uid"))
	is.Equal("1", c.Param("fid"))
}

func TestRouterParamWithSlash(t *testing.T) {
	g := New()
	r := g.router

	r.Add("/a/:b/c/d/:g", func(c Context) error {
		return nil
	})

	r.Add("/a/:b/c/:d/:f", func(c Context) error {
		return nil
	})

	c := g.NewContext(nil, nil, "", nil).(*context)

	// No Panic
	r.Find("/a/1/c/d/2/3", c)
}

func TestRouterParamStaticConflict(t *testing.T) {
	g := New()
	r := g.router
	handler := func(c Context) error {
		c.Set("path", c.Path())
		return nil
	}

	gr := g.Group("/g")
	gr.Handle("/skills", handler)
	gr.Handle("/status", handler)
	gr.Handle("/:name", handler)

	is := is.New(t)

	c := g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/g/s", c)

	is.NoErr(c.handler(c))
	is.Equal("s", c.Param("name"))
	is.Equal("/g/:name", c.Get("path"))

	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/g/status", c)
	is.NoErr(c.handler(c))
	is.Equal("/g/status", c.Get("path"))
}

func TestRouterMatchAny(t *testing.T) {
	g := New()
	r := g.router

	// Routes
	r.Add("/", func(Context) error {
		return nil
	})
	r.Add("/*", func(Context) error {
		return nil
	})
	r.Add("/users/*", func(Context) error {
		return nil
	})
	c := g.NewContext(nil, nil, "", nil).(*context)

	is := is.New(t)

	r.Find("/", c)
	is.Equal("", c.Param("*"))

	r.Find("/download", c)
	is.Equal("download", c.Param("*"))

	r.Find("/users/joe", c)
	is.Equal("joe", c.Param("*"))
}

// TestRouterMatchAnySlash shall verify finding the best route
// for any routes with trailing slash requests
func TestRouterMatchAnySlash(t *testing.T) {
	g := New()
	r := g.router

	is := is.New(t)

	handler := func(c Context) error {
		c.Set("path", c.Path())
		return nil
	}

	// Routes
	r.Add("/users", handler)
	r.Add("/users/*", handler)
	r.Add("/img/*", handler)
	r.Add("/img/load", handler)
	r.Add("/img/load/*", handler)
	r.Add("/assets/*", handler)

	c := g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/", c)
	is.Equal("", c.Param("*"))

	// Test trailing slash request for simple any route (see #1526)
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/", c)
	is.NoErr(c.handler(c))
	is.Equal("/users/*", c.Get("path"))
	is.Equal("", c.Param("*"))

	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/joe", c)
	is.NoErr(c.handler(c))
	is.Equal("/users/*", c.Get("path"))
	is.Equal("joe", c.Param("*"))

	// Test trailing slash request for nested any route (see #1526)
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/img/load", c)
	is.NoErr(c.handler(c))
	is.Equal("/img/load", c.Get("path"))
	is.Equal("", c.Param("*"))

	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/img/load/", c)
	is.NoErr(c.handler(c))
	is.Equal("/img/load/*", c.Get("path"))
	is.Equal("", c.Param("*"))

	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/img/load/ben", c)
	is.NoErr(c.handler(c))
	is.Equal("/img/load/*", c.Get("path"))
	is.Equal("ben", c.Param("*"))

	// Test /assets/* any route
	// ... without trailing slash must not match
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/assets", c)
	is.True(c.handler(c) != nil)
	is.Equal(nil, c.Get("path"))
	is.Equal("", c.Param("*"))

	// ... with trailing slash must match
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/assets/", c)
	is.NoErr(c.handler(c))
	is.Equal("/assets/*", c.Get("path"))
	is.Equal("", c.Param("*"))

}

func TestRouterMatchAnyMultiLevel(t *testing.T) {
	is := is.New(t)

	g := New()
	r := g.router
	handler := func(c Context) error {
		c.Set("path", c.Path())
		return nil
	}

	// Routes
	r.Add("/api/users/jack", handler)
	r.Add("/api/users/jill", handler)
	r.Add("/api/users/*", handler)
	r.Add("/api/*", handler)
	r.Add("/other/*", handler)
	r.Add("/*", handler)

	c := g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/api/users/jack", c)
	is.NoErr(c.handler(c))
	is.Equal("/api/users/jack", c.Get("path"))
	is.Equal("", c.Param("*"))

	r.Find("/api/users/jill", c)
	is.NoErr(c.handler(c))
	is.Equal("/api/users/jill", c.Get("path"))
	is.Equal("", c.Param("*"))

	r.Find("/api/users/joe", c)
	is.NoErr(c.handler(c))
	is.Equal("/api/users/*", c.Get("path"))
	is.Equal("joe", c.Param("*"))

	r.Find("/api/nousers/joe", c)
	is.NoErr(c.handler(c))
	is.Equal("/api/*", c.Get("path"))
	is.Equal("nousers/joe", c.Param("*"))

	r.Find("/api/none", c)
	is.NoErr(c.handler(c))
	is.Equal("/api/*", c.Get("path"))
	is.Equal("none", c.Param("*"))

	r.Find("/noapi/users/jim", c)
	is.NoErr(c.handler(c))
	is.Equal("/*", c.Get("path"))
	is.Equal("noapi/users/jim", c.Param("*"))
}
func TestRouterMatchAnyMultiLevelWithPost(t *testing.T) {
	is := is.New(t)

	g := New()
	r := g.router
	handler := func(c Context) error {
		c.Set("path", c.Path())
		return nil
	}

	// Routes
	g.Handle("/api/auth/login", handler)
	g.Handle("/api/auth/forgotPassword", handler)
	g.Handle("/api/*", handler)
	g.Handle("/*", handler)

	// /api/auth/login shall choose login
	c := g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/api/auth/login", c)
	is.NoErr(c.handler(c))
	is.Equal("/api/auth/login", c.Get("path"))
	is.Equal("", c.Param("*"))

	// /api/auth/login shall choose any route
	// c = g.NewContext(nil, nil,nil).(*context)
	// r.Find( "/api/auth/login", c)
	// c.handler(c)
	// is.Equal("/api/*", c.Get("path"))
	// is.Equal("auth/login", c.Param("*"))

	// /api/auth/logout shall choose nearest any route
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/api/auth/logout", c)
	is.NoErr(c.handler(c))
	is.Equal("/api/*", c.Get("path"))
	is.Equal("auth/logout", c.Param("*"))

	// /api/other/test shall choose nearest any route
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/api/other/test", c)
	is.NoErr(c.handler(c))
	is.Equal("/api/*", c.Get("path"))
	is.Equal("other/test", c.Param("*"))

	// /api/other/test shall choose nearest any route
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/api/other/test", c)
	is.NoErr(c.handler(c))
	is.Equal("/api/*", c.Get("path"))
	is.Equal("other/test", c.Param("*"))

}

func TestRouterMicroParam(t *testing.T) {
	is := is.New(t)

	g := New()
	r := g.router
	r.Add("/:a/:b/:c", func(c Context) error {
		return nil
	})
	c := g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/1/2/3", c)
	is.Equal("1", c.Param("a"))
	is.Equal("2", c.Param("b"))
	is.Equal("3", c.Param("c"))
}

func TestRouterMixParamMatchAny(t *testing.T) {
	is := is.New(t)

	g := New()
	r := g.router

	// Route
	r.Add("/users/:id/*", func(c Context) error {
		return nil
	})
	c := g.NewContext(nil, nil, "", nil).(*context)

	r.Find("/users/joe/comments", c)
	is.NoErr(c.handler(c))
	is.Equal("joe", c.Param("id"))
}

func TestRouterMultiRoute(t *testing.T) {
	is := is.New(t)

	g := New()
	r := g.router

	// Routes
	r.Add("/users", func(c Context) error {
		c.Set("path", "/users")
		return nil
	})
	r.Add("/users/:id", func(c Context) error {
		return nil
	})
	c := g.NewContext(nil, nil, "", nil).(*context)

	// Route > /users
	r.Find("/users", c)
	is.NoErr(c.handler(c))
	is.Equal("/users", c.Get("path"))

	// Route > /users/:id
	r.Find("/users/1", c)
	is.Equal("1", c.Param("id"))

	// Route > /user
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/user", c)
	he := c.handler(c).(*GeminiError)
	is.Equal(StatusNotFound, he.Code)
}

func TestRouterPriority(t *testing.T) {
	is := is.New(t)

	g := New()
	r := g.router

	// Routes
	r.Add("/users", handlerHelper("a", 1))
	r.Add("/users/new", handlerHelper("b", 2))
	r.Add("/users/:id", handlerHelper("c", 3))
	r.Add("/users/dew", handlerHelper("d", 4))
	r.Add("/users/:id/files", handlerHelper("g", 5))
	r.Add("/users/newsee", handlerHelper("f", 6))
	r.Add("/users/*", handlerHelper("g", 7))
	r.Add("/users/new/*", handlerHelper("h", 8))
	r.Add("/*", handlerHelper("i", 9))

	// Route > /users
	c := g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users", c)
	is.NoErr(c.handler(c))
	is.Equal(1, c.Get("a"))
	is.Equal("/users", c.Get("path"))

	// Route > /users/new
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/new", c)
	is.NoErr(c.handler(c))
	is.Equal(2, c.Get("b"))
	is.Equal("/users/new", c.Get("path"))

	// Route > /users/:id
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/1", c)
	is.NoErr(c.handler(c))
	is.Equal(3, c.Get("c"))
	is.Equal("/users/:id", c.Get("path"))

	// Route > /users/dew
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/dew", c)
	is.NoErr(c.handler(c))
	is.Equal(4, c.Get("d"))
	is.Equal("/users/dew", c.Get("path"))

	// Route > /users/:id/files
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/1/files", c)
	is.NoErr(c.handler(c))
	is.Equal(5, c.Get("g"))
	is.Equal("/users/:id/files", c.Get("path"))

	// Route > /users/:id
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/news", c)
	is.NoErr(c.handler(c))
	is.Equal(3, c.Get("c"))
	is.Equal("/users/:id", c.Get("path"))

	// Route > /users/newsee
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/newsee", c)
	is.NoErr(c.handler(c))
	is.Equal(6, c.Get("f"))
	is.Equal("/users/newsee", c.Get("path"))

	// Route > /users/newsee
	r.Find("/users/newsee", c)
	is.NoErr(c.handler(c))
	is.Equal(6, c.Get("f"))

	// Route > /users/newsee
	r.Find("/users/newsee", c)
	is.NoErr(c.handler(c))
	is.Equal(6, c.Get("f"))

	// Route > /users/*
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/joe/books", c)
	is.NoErr(c.handler(c))
	is.Equal(7, c.Get("g"))
	is.Equal("/users/*", c.Get("path"))
	is.Equal("joe/books", c.Param("*"))

	// Route > /users/new/* should be matched
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/new/someone", c)
	is.NoErr(c.handler(c))
	is.Equal(8, c.Get("h"))
	is.Equal("/users/new/*", c.Get("path"))
	is.Equal("someone", c.Param("*"))

	// Route > /users/* should be matched although /users/dew exists
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/dew/someone", c)
	is.NoErr(c.handler(c))
	is.Equal(7, c.Get("g"))
	is.Equal("/users/*", c.Get("path"))

	is.Equal("dew/someone", c.Param("*"))

	// Route > /users/* should be matched although /users/dew exists
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/notexists/someone", c)
	is.NoErr(c.handler(c))
	is.Equal(7, c.Get("g"))
	is.Equal("/users/*", c.Get("path"))
	is.Equal("notexists/someone", c.Param("*"))

	// Route > *
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/nousers", c)
	is.NoErr(c.handler(c))
	is.Equal(9, c.Get("i"))
	is.Equal("/*", c.Get("path"))
	is.Equal("nousers", c.Param("*"))

	// Route > *
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/nousers/new", c)
	is.NoErr(c.handler(c))
	is.Equal(9, c.Get("i"))
	is.Equal("/*", c.Get("path"))
	is.Equal("nousers/new", c.Param("*"))
}

func TestRouterIssue1348(t *testing.T) {
	g := New()
	r := g.router

	r.Add("/:lang/", func(c Context) error {
		return nil
	})
	r.Add("/:lang/dupa", func(c Context) error {
		return nil
	})
}

func TestRouterPriorityNotFound(t *testing.T) {
	is := is.New(t)

	g := New()
	r := g.router
	c := g.NewContext(nil, nil, "", nil).(*context)

	// Add
	r.Add("/a/foo", func(c Context) error {
		c.Set("a", 1)
		return nil
	})
	r.Add("/a/bar", func(c Context) error {
		c.Set("b", 2)
		return nil
	})

	// Find
	r.Find("/a/foo", c)
	is.NoErr(c.handler(c))
	is.Equal(1, c.Get("a"))

	r.Find("/a/bar", c)
	is.NoErr(c.handler(c))
	is.Equal(2, c.Get("b"))

	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/abc/def", c)
	he := c.handler(c).(*GeminiError)
	is.Equal(StatusNotFound, he.Code)
}

func TestRouterParamNames(t *testing.T) {
	is := is.New(t)

	g := New()
	r := g.router

	// Routes
	r.Add("/users", func(c Context) error {
		c.Set("path", "/users")
		return nil
	})
	r.Add("/users/:id", func(c Context) error {
		return nil
	})
	r.Add("/users/:uid/files/:fid", func(c Context) error {
		return nil
	})
	c := g.NewContext(nil, nil, "", nil).(*context)

	// Route > /users
	r.Find("/users", c)
	is.NoErr(c.handler(c))
	is.Equal("/users", c.Get("path"))

	// Route > /users/:id
	r.Find("/users/1", c)
	is.Equal("id", c.pnames[0])
	is.Equal("1", c.Param("id"))

	// Route > /users/:uid/files/:fid
	r.Find("/users/1/files/1", c)
	is.Equal("uid", c.pnames[0])
	is.Equal("1", c.Param("uid"))
	is.Equal("fid", c.pnames[1])
	is.Equal("1", c.Param("fid"))
}

func TestRouterStaticDynamicConflict(t *testing.T) {
	is := is.New(t)

	g := New()
	r := g.router

	r.Add("/dictionary/skills", handlerHelper("a", 1))
	r.Add("/dictionary/:name", handlerHelper("b", 2))
	r.Add("/users/new", handlerHelper("d", 4))
	r.Add("/users/:name", handlerHelper("g", 5))
	r.Add("/server", handlerHelper("c", 3))
	r.Add("/", handlerHelper("f", 6))

	c := g.NewContext(nil, nil, "", nil)
	r.Find("/dictionary/skills", c)
	is.NoErr(c.Handler()(c))
	is.Equal(1, c.Get("a"))
	is.Equal("/dictionary/skills", c.Get("path"))

	c = g.NewContext(nil, nil, "", nil)
	r.Find("/dictionary/skillsnot", c)
	is.NoErr(c.Handler()(c))
	is.Equal(2, c.Get("b"))
	is.Equal("/dictionary/:name", c.Get("path"))

	c = g.NewContext(nil, nil, "", nil)
	r.Find("/dictionary/type", c)
	is.NoErr(c.Handler()(c))
	is.Equal(2, c.Get("b"))
	is.Equal("/dictionary/:name", c.Get("path"))

	c = g.NewContext(nil, nil, "", nil)
	r.Find("/server", c)
	is.NoErr(c.Handler()(c))
	is.Equal(3, c.Get("c"))
	is.Equal("/server", c.Get("path"))

	c = g.NewContext(nil, nil, "", nil)
	r.Find("/users/new", c)
	is.NoErr(c.Handler()(c))
	is.Equal(4, c.Get("d"))
	is.Equal("/users/new", c.Get("path"))

	c = g.NewContext(nil, nil, "", nil)
	r.Find("/users/new2", c)
	is.NoErr(c.Handler()(c))
	is.Equal(5, c.Get("g"))
	is.Equal("/users/:name", c.Get("path"))

	c = g.NewContext(nil, nil, "", nil)
	r.Find("/", c)
	is.NoErr(c.Handler()(c))
	is.Equal(6, c.Get("f"))
	is.Equal("/", c.Get("path"))
}

func TestRouterParamBacktraceNotFound(t *testing.T) {
	is := is.New(t)

	g := New()
	r := g.router

	// Add
	r.Add("/:param1", func(c Context) error {
		return nil
	})
	r.Add("/:param1/foo", func(c Context) error {
		return nil
	})
	r.Add("/:param1/bar", func(c Context) error {
		return nil
	})
	r.Add("/:param1/bar/:param2", func(c Context) error {
		return nil
	})

	c := g.NewContext(nil, nil, "", nil).(*context)

	//Find
	r.Find("/a", c)
	is.Equal("a", c.Param("param1"))

	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/a/foo", c)
	is.Equal("a", c.Param("param1"))

	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/a/bar", c)
	is.Equal("a", c.Param("param1"))

	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/a/bar/b", c)
	is.Equal("a", c.Param("param1"))
	is.Equal("b", c.Param("param2"))

	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/a/bbbbb", c)
	he := c.handler(c).(*GeminiError)
	is.Equal(StatusNotFound, he.Code)
}

func testRouterAPI(t *testing.T, api []*Route) {
	is := is.New(t)

	g := New()
	r := g.router

	for _, route := range api {
		r.Add(route.Path, func(c Context) error {
			return nil
		})
	}
	c := g.NewContext(nil, nil, "", nil).(*context)
	for _, route := range api {
		r.Find(route.Path, c)
		tokens := strings.Split(route.Path[1:], "/")
		for _, token := range tokens {
			if token[0] == ':' {
				is.Equal(c.Param(token[1:]), token)
			}
		}
	}
}

func TestRouterGitHubAPI(t *testing.T) {
	testRouterAPI(t, gitHubAPI)
}

func TestRouterParamAlias(t *testing.T) {
	api := []*Route{
		{"/users/:userID/following", ""},
		{"/users/:userID/followedBy", ""},
		{"/users/:userID/follow", ""},
	}
	testRouterAPI(t, api)
}

func TestRouterParamOrdering(t *testing.T) {
	api := []*Route{
		{"/:a/:b/:c/:id", ""},
		{"/:a/:id", ""},
		{"/:a/:g/:id", ""},
	}
	testRouterAPI(t, api)
	api2 := []*Route{
		{"/:a/:id", ""},
		{"/:a/:g/:id", ""},
		{"/:a/:b/:c/:id", ""},
	}
	testRouterAPI(t, api2)
	api3 := []*Route{
		{"/:a/:b/:c/:id", ""},
		{"/:a/:g/:id", ""},
		{"/:a/:id", ""},
	}
	testRouterAPI(t, api3)
}

func TestRouterMixedParams(t *testing.T) {
	api := []*Route{
		{"/teacher/:tid/room/suggestions", ""},
		{"/teacher/:id", ""},
	}
	testRouterAPI(t, api)
	api2 := []*Route{
		{"/teacher/:id", ""},
		{"/teacher/:tid/room/suggestions", ""},
	}
	testRouterAPI(t, api2)
}

func TestRouterParam1466(t *testing.T) {
	is := is.New(t)

	g := New()
	r := g.router

	r.Add("/users/signup", func(c Context) error {
		return nil
	})
	r.Add("/users/signup/bulk", func(c Context) error {
		return nil
	})
	r.Add("/users/survey", func(c Context) error {
		return nil
	})
	r.Add("/users/:username", func(c Context) error {
		return nil
	})
	r.Add("/interests/:name/users", func(c Context) error {
		return nil
	})
	r.Add("/skills/:name/users", func(c Context) error {
		return nil
	})
	// Additional routes for Issue 1479
	r.Add("/users/:username/likes/projects/ids", func(c Context) error {
		return nil
	})
	r.Add("/users/:username/profile", func(c Context) error {
		return nil
	})
	r.Add("/users/:username/uploads/:type", func(c Context) error {
		return nil
	})

	c := g.NewContext(nil, nil, "", nil).(*context)

	r.Find("/users/ajitem", c)
	is.Equal("ajitem", c.Param("username"))

	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/sharewithme", c)
	is.Equal("sharewithme", c.Param("username"))

	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/signup", c)
	is.Equal("", c.Param("username"))
	// Additional assertions for #1479
	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/sharewithme/likes/projects/ids", c)
	is.Equal("sharewithme", c.Param("username"))

	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/ajitem/likes/projects/ids", c)
	is.Equal("ajitem", c.Param("username"))

	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/sharewithme/profile", c)
	is.Equal("sharewithme", c.Param("username"))

	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/ajitem/profile", c)
	is.Equal("ajitem", c.Param("username"))

	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/sharewithme/uploads/self", c)
	is.Equal("sharewithme", c.Param("username"))
	is.Equal("self", c.Param("type"))

	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/ajitem/uploads/self", c)
	is.Equal("ajitem", c.Param("username"))
	is.Equal("self", c.Param("type"))

	c = g.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/tree/free", c)
	is.Equal("", c.Param("id"))
	is.Equal(Status(0), c.response.Status)
}

func benchmarkRouterRoutes(b *testing.B, routes []*Route) {
	g := New()
	r := g.router
	b.ReportAllocs()

	// Add routes
	for _, route := range routes {
		r.Add(route.Path, func(c Context) error {
			return nil
		})
	}

	// Find routes
	for i := 0; i < b.N; i++ {
		for _, route := range gitHubAPI {
			c := g.pool.Get().(*context)
			r.Find(route.Path, c)
			g.pool.Put(c)
		}
	}
}

func BenchmarkRouterStaticRoutes(b *testing.B) {
	benchmarkRouterRoutes(b, staticRoutes)
}

func BenchmarkRouterGitHubAPI(b *testing.B) {
	benchmarkRouterRoutes(b, gitHubAPI)
}

func BenchmarkRouterParseAPI(b *testing.B) {
	benchmarkRouterRoutes(b, parseAPI)
}

func BenchmarkRouterGooglePlusAPI(b *testing.B) {
	benchmarkRouterRoutes(b, googlePlusAPI)
}
