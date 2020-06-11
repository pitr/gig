package gig

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
	e := New()
	r := e.router
	path := ""
	r.Add(path, func(c Context) error {
		c.Set("path", path)
		return nil
	})
	c := e.NewContext(nil, nil, "", nil).(*context)
	r.Find(path, c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, path, c.Get("path"))
}

func TestRouterStatic(t *testing.T) {
	e := New()
	r := e.router
	path := "/folders/a/files/gig.gif"
	r.Add(path, func(c Context) error {
		c.Set("path", path)
		return nil
	})
	c := e.NewContext(nil, nil, "", nil).(*context)
	r.Find(path, c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, path, c.Get("path"))
}

func TestRouterParam(t *testing.T) {
	e := New()
	r := e.router
	r.Add("/users/:id", func(c Context) error {
		return nil
	})
	c := e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/1", c)
	assert.Equal(t, "1", c.Param("id"))
}

func TestRouterTwoParam(t *testing.T) {
	e := New()
	r := e.router
	r.Add("/users/:uid/files/:fid", func(Context) error {
		return nil
	})
	c := e.NewContext(nil, nil, "", nil).(*context)

	r.Find("/users/1/files/1", c)
	assert.Equal(t, "1", c.Param("uid"))
	assert.Equal(t, "1", c.Param("fid"))
}

// Issue #378
func TestRouterParamWithSlash(t *testing.T) {
	e := New()
	r := e.router

	r.Add("/a/:b/c/d/:e", func(c Context) error {
		return nil
	})

	r.Add("/a/:b/c/:d/:f", func(c Context) error {
		return nil
	})

	c := e.NewContext(nil, nil, "", nil).(*context)
	assert.NotPanics(t, func() {
		r.Find("/a/1/c/d/2/3", c)
	})
}

// Issue #1509
func TestRouterParamStaticConflict(t *testing.T) {
	e := New()
	r := e.router
	handler := func(c Context) error {
		c.Set("path", c.Path())
		return nil
	}

	g := e.Group("/g")
	g.Handle("/skills", handler)
	g.Handle("/status", handler)
	g.Handle("/:name", handler)

	c := e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/g/s", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "s", c.Param("name"))
	assert.Equal(t, "/g/:name", c.Get("path"))

	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/g/status", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "/g/status", c.Get("path"))
}

func TestRouterMatchAny(t *testing.T) {
	e := New()
	r := e.router

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
	c := e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/", c)
	assert.Equal(t, "", c.Param("*"))

	r.Find("/download", c)
	assert.Equal(t, "download", c.Param("*"))

	r.Find("/users/joe", c)
	assert.Equal(t, "joe", c.Param("*"))
}

// TestRouterMatchAnySlash shall verify finding the best route
// for any routes with trailing slash requests
func TestRouterMatchAnySlash(t *testing.T) {
	e := New()
	r := e.router

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

	c := e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/", c)
	assert.Equal(t, "", c.Param("*"))

	// Test trailing slash request for simple any route (see #1526)
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "/users/*", c.Get("path"))
	assert.Equal(t, "", c.Param("*"))

	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/joe", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "/users/*", c.Get("path"))
	assert.Equal(t, "joe", c.Param("*"))

	// Test trailing slash request for nested any route (see #1526)
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/img/load", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "/img/load", c.Get("path"))
	assert.Equal(t, "", c.Param("*"))

	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/img/load/", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "/img/load/*", c.Get("path"))
	assert.Equal(t, "", c.Param("*"))

	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/img/load/ben", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "/img/load/*", c.Get("path"))
	assert.Equal(t, "ben", c.Param("*"))

	// Test /assets/* any route
	// ... without trailing slash must not match
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/assets", c)
	assert.Error(t, c.handler(c))
	assert.Equal(t, nil, c.Get("path"))
	assert.Equal(t, "", c.Param("*"))

	// ... with trailing slash must match
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/assets/", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "/assets/*", c.Get("path"))
	assert.Equal(t, "", c.Param("*"))

}

func TestRouterMatchAnyMultiLevel(t *testing.T) {
	e := New()
	r := e.router
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

	c := e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/api/users/jack", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "/api/users/jack", c.Get("path"))
	assert.Equal(t, "", c.Param("*"))

	r.Find("/api/users/jill", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "/api/users/jill", c.Get("path"))
	assert.Equal(t, "", c.Param("*"))

	r.Find("/api/users/joe", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "/api/users/*", c.Get("path"))
	assert.Equal(t, "joe", c.Param("*"))

	r.Find("/api/nousers/joe", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "/api/*", c.Get("path"))
	assert.Equal(t, "nousers/joe", c.Param("*"))

	r.Find("/api/none", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "/api/*", c.Get("path"))
	assert.Equal(t, "none", c.Param("*"))

	r.Find("/noapi/users/jim", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "/*", c.Get("path"))
	assert.Equal(t, "noapi/users/jim", c.Param("*"))
}
func TestRouterMatchAnyMultiLevelWithPost(t *testing.T) {
	e := New()
	r := e.router
	handler := func(c Context) error {
		c.Set("path", c.Path())
		return nil
	}

	// Routes
	e.Handle("/api/auth/login", handler)
	e.Handle("/api/auth/forgotPassword", handler)
	e.Handle("/api/*", handler)
	e.Handle("/*", handler)

	// /api/auth/login shall choose login
	c := e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/api/auth/login", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "/api/auth/login", c.Get("path"))
	assert.Equal(t, "", c.Param("*"))

	// /api/auth/login shall choose any route
	// c = e.NewContext(nil, nil,nil).(*context)
	// r.Find( "/api/auth/login", c)
	// c.handler(c)
	// assert.Equal(t, "/api/*", c.Get("path"))
	// assert.Equal(t, "auth/login", c.Param("*"))

	// /api/auth/logout shall choose nearest any route
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/api/auth/logout", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "/api/*", c.Get("path"))
	assert.Equal(t, "auth/logout", c.Param("*"))

	// /api/other/test shall choose nearest any route
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/api/other/test", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "/api/*", c.Get("path"))
	assert.Equal(t, "other/test", c.Param("*"))

	// /api/other/test shall choose nearest any route
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/api/other/test", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "/api/*", c.Get("path"))
	assert.Equal(t, "other/test", c.Param("*"))

}

func TestRouterMicroParam(t *testing.T) {
	e := New()
	r := e.router
	r.Add("/:a/:b/:c", func(c Context) error {
		return nil
	})
	c := e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/1/2/3", c)
	assert.Equal(t, "1", c.Param("a"))
	assert.Equal(t, "2", c.Param("b"))
	assert.Equal(t, "3", c.Param("c"))
}

func TestRouterMixParamMatchAny(t *testing.T) {
	e := New()
	r := e.router

	// Route
	r.Add("/users/:id/*", func(c Context) error {
		return nil
	})
	c := e.NewContext(nil, nil, "", nil).(*context)

	r.Find("/users/joe/comments", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "joe", c.Param("id"))
}

func TestRouterMultiRoute(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add("/users", func(c Context) error {
		c.Set("path", "/users")
		return nil
	})
	r.Add("/users/:id", func(c Context) error {
		return nil
	})
	c := e.NewContext(nil, nil, "", nil).(*context)

	// Route > /users
	r.Find("/users", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "/users", c.Get("path"))

	// Route > /users/:id
	r.Find("/users/1", c)
	assert.Equal(t, "1", c.Param("id"))

	// Route > /user
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/user", c)
	he := c.handler(c).(*GeminiError)
	assert.Equal(t, StatusNotFound, he.Code)
}

func TestRouterPriority(t *testing.T) {
	e := New()
	r := e.router

	// Routes
	r.Add("/users", handlerHelper("a", 1))
	r.Add("/users/new", handlerHelper("b", 2))
	r.Add("/users/:id", handlerHelper("c", 3))
	r.Add("/users/dew", handlerHelper("d", 4))
	r.Add("/users/:id/files", handlerHelper("e", 5))
	r.Add("/users/newsee", handlerHelper("f", 6))
	r.Add("/users/*", handlerHelper("g", 7))
	r.Add("/users/new/*", handlerHelper("h", 8))
	r.Add("/*", handlerHelper("i", 9))

	// Route > /users
	c := e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, 1, c.Get("a"))
	assert.Equal(t, "/users", c.Get("path"))

	// Route > /users/new
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/new", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, 2, c.Get("b"))
	assert.Equal(t, "/users/new", c.Get("path"))

	// Route > /users/:id
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/1", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, 3, c.Get("c"))
	assert.Equal(t, "/users/:id", c.Get("path"))

	// Route > /users/dew
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/dew", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, 4, c.Get("d"))
	assert.Equal(t, "/users/dew", c.Get("path"))

	// Route > /users/:id/files
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/1/files", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, 5, c.Get("e"))
	assert.Equal(t, "/users/:id/files", c.Get("path"))

	// Route > /users/:id
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/news", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, 3, c.Get("c"))
	assert.Equal(t, "/users/:id", c.Get("path"))

	// Route > /users/newsee
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/newsee", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, 6, c.Get("f"))
	assert.Equal(t, "/users/newsee", c.Get("path"))

	// Route > /users/newsee
	r.Find("/users/newsee", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, 6, c.Get("f"))

	// Route > /users/newsee
	r.Find("/users/newsee", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, 6, c.Get("f"))

	// Route > /users/*
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/joe/books", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, 7, c.Get("g"))
	assert.Equal(t, "/users/*", c.Get("path"))
	assert.Equal(t, "joe/books", c.Param("*"))

	// Route > /users/new/* should be matched
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/new/someone", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, 8, c.Get("h"))
	assert.Equal(t, "/users/new/*", c.Get("path"))
	assert.Equal(t, "someone", c.Param("*"))

	// Route > /users/* should be matched although /users/dew exists
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/dew/someone", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, 7, c.Get("g"))
	assert.Equal(t, "/users/*", c.Get("path"))

	assert.Equal(t, "dew/someone", c.Param("*"))

	// Route > /users/* should be matched although /users/dew exists
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/notexists/someone", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, 7, c.Get("g"))
	assert.Equal(t, "/users/*", c.Get("path"))
	assert.Equal(t, "notexists/someone", c.Param("*"))

	// Route > *
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/nousers", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, 9, c.Get("i"))
	assert.Equal(t, "/*", c.Get("path"))
	assert.Equal(t, "nousers", c.Param("*"))

	// Route > *
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/nousers/new", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, 9, c.Get("i"))
	assert.Equal(t, "/*", c.Get("path"))
	assert.Equal(t, "nousers/new", c.Param("*"))
}

func TestRouterIssue1348(t *testing.T) {
	e := New()
	r := e.router

	r.Add("/:lang/", func(c Context) error {
		return nil
	})
	r.Add("/:lang/dupa", func(c Context) error {
		return nil
	})
}

// Issue #372
func TestRouterPriorityNotFound(t *testing.T) {
	e := New()
	r := e.router
	c := e.NewContext(nil, nil, "", nil).(*context)

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
	assert.NoError(t, c.handler(c))
	assert.Equal(t, 1, c.Get("a"))

	r.Find("/a/bar", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, 2, c.Get("b"))

	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/abc/def", c)
	he := c.handler(c).(*GeminiError)
	assert.Equal(t, StatusNotFound, he.Code)
}

func TestRouterParamNames(t *testing.T) {
	e := New()
	r := e.router

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
	c := e.NewContext(nil, nil, "", nil).(*context)

	// Route > /users
	r.Find("/users", c)
	assert.NoError(t, c.handler(c))
	assert.Equal(t, "/users", c.Get("path"))

	// Route > /users/:id
	r.Find("/users/1", c)
	assert.Equal(t, "id", c.pnames[0])
	assert.Equal(t, "1", c.Param("id"))

	// Route > /users/:uid/files/:fid
	r.Find("/users/1/files/1", c)
	assert.Equal(t, "uid", c.pnames[0])
	assert.Equal(t, "1", c.Param("uid"))
	assert.Equal(t, "fid", c.pnames[1])
	assert.Equal(t, "1", c.Param("fid"))
}

// Issue #623 and #1406
func TestRouterStaticDynamicConflict(t *testing.T) {
	e := New()
	r := e.router

	r.Add("/dictionary/skills", handlerHelper("a", 1))
	r.Add("/dictionary/:name", handlerHelper("b", 2))
	r.Add("/users/new", handlerHelper("d", 4))
	r.Add("/users/:name", handlerHelper("e", 5))
	r.Add("/server", handlerHelper("c", 3))
	r.Add("/", handlerHelper("f", 6))

	c := e.NewContext(nil, nil, "", nil)
	r.Find("/dictionary/skills", c)
	assert.NoError(t, c.Handler()(c))
	assert.Equal(t, 1, c.Get("a"))
	assert.Equal(t, "/dictionary/skills", c.Get("path"))

	c = e.NewContext(nil, nil, "", nil)
	r.Find("/dictionary/skillsnot", c)
	assert.NoError(t, c.Handler()(c))
	assert.Equal(t, 2, c.Get("b"))
	assert.Equal(t, "/dictionary/:name", c.Get("path"))

	c = e.NewContext(nil, nil, "", nil)
	r.Find("/dictionary/type", c)
	assert.NoError(t, c.Handler()(c))
	assert.Equal(t, 2, c.Get("b"))
	assert.Equal(t, "/dictionary/:name", c.Get("path"))

	c = e.NewContext(nil, nil, "", nil)
	r.Find("/server", c)
	assert.NoError(t, c.Handler()(c))
	assert.Equal(t, 3, c.Get("c"))
	assert.Equal(t, "/server", c.Get("path"))

	c = e.NewContext(nil, nil, "", nil)
	r.Find("/users/new", c)
	assert.NoError(t, c.Handler()(c))
	assert.Equal(t, 4, c.Get("d"))
	assert.Equal(t, "/users/new", c.Get("path"))

	c = e.NewContext(nil, nil, "", nil)
	r.Find("/users/new2", c)
	assert.NoError(t, c.Handler()(c))
	assert.Equal(t, 5, c.Get("e"))
	assert.Equal(t, "/users/:name", c.Get("path"))

	c = e.NewContext(nil, nil, "", nil)
	r.Find("/", c)
	assert.NoError(t, c.Handler()(c))
	assert.Equal(t, 6, c.Get("f"))
	assert.Equal(t, "/", c.Get("path"))
}

// Issue #1348
func TestRouterParamBacktraceNotFound(t *testing.T) {
	e := New()
	r := e.router

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

	c := e.NewContext(nil, nil, "", nil).(*context)

	//Find
	r.Find("/a", c)
	assert.Equal(t, "a", c.Param("param1"))

	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/a/foo", c)
	assert.Equal(t, "a", c.Param("param1"))

	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/a/bar", c)
	assert.Equal(t, "a", c.Param("param1"))

	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/a/bar/b", c)
	assert.Equal(t, "a", c.Param("param1"))
	assert.Equal(t, "b", c.Param("param2"))

	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/a/bbbbb", c)
	he := c.handler(c).(*GeminiError)
	assert.Equal(t, StatusNotFound, he.Code)
}

func testRouterAPI(t *testing.T, api []*Route) {
	e := New()
	r := e.router

	for _, route := range api {
		r.Add(route.Path, func(c Context) error {
			return nil
		})
	}
	c := e.NewContext(nil, nil, "", nil).(*context)
	for _, route := range api {
		r.Find(route.Path, c)
		tokens := strings.Split(route.Path[1:], "/")
		for _, token := range tokens {
			if token[0] == ':' {
				assert.Equal(t, c.Param(token[1:]), token)
			}
		}
	}
}

func TestRouterGitHubAPI(t *testing.T) {
	testRouterAPI(t, gitHubAPI)
}

// Issue #729
func TestRouterParamAlias(t *testing.T) {
	api := []*Route{
		{"/users/:userID/following", ""},
		{"/users/:userID/followedBy", ""},
		{"/users/:userID/follow", ""},
	}
	testRouterAPI(t, api)
}

// Issue #1052
func TestRouterParamOrdering(t *testing.T) {
	api := []*Route{
		{"/:a/:b/:c/:id", ""},
		{"/:a/:id", ""},
		{"/:a/:e/:id", ""},
	}
	testRouterAPI(t, api)
	api2 := []*Route{
		{"/:a/:id", ""},
		{"/:a/:e/:id", ""},
		{"/:a/:b/:c/:id", ""},
	}
	testRouterAPI(t, api2)
	api3 := []*Route{
		{"/:a/:b/:c/:id", ""},
		{"/:a/:e/:id", ""},
		{"/:a/:id", ""},
	}
	testRouterAPI(t, api3)
}

// Issue #1139
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

// Issue #1466
func TestRouterParam1466(t *testing.T) {
	e := New()
	r := e.router

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

	c := e.NewContext(nil, nil, "", nil).(*context)

	r.Find("/users/ajitem", c)
	assert.Equal(t, "ajitem", c.Param("username"))

	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/sharewithme", c)
	assert.Equal(t, "sharewithme", c.Param("username"))

	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/signup", c)
	assert.Equal(t, "", c.Param("username"))
	// Additional assertions for #1479
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/sharewithme/likes/projects/ids", c)
	assert.Equal(t, "sharewithme", c.Param("username"))

	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/ajitem/likes/projects/ids", c)
	assert.Equal(t, "ajitem", c.Param("username"))

	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/sharewithme/profile", c)
	assert.Equal(t, "sharewithme", c.Param("username"))

	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/ajitem/profile", c)
	assert.Equal(t, "ajitem", c.Param("username"))

	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/sharewithme/uploads/self", c)
	assert.Equal(t, "sharewithme", c.Param("username"))
	assert.Equal(t, "self", c.Param("type"))

	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/ajitem/uploads/self", c)
	assert.Equal(t, "ajitem", c.Param("username"))
	assert.Equal(t, "self", c.Param("type"))

	// Issue #1493 - check for routing loop
	c = e.NewContext(nil, nil, "", nil).(*context)
	r.Find("/users/tree/free", c)
	assert.Equal(t, "", c.Param("id"))
	assert.Equal(t, Status(0), c.response.Status)
}

func benchmarkRouterRoutes(b *testing.B, routes []*Route) {
	e := New()
	r := e.router
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
			c := e.pool.Get().(*context)
			r.Find(route.Path, c)
			e.pool.Put(c)
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
