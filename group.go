package gig

type (
	// Group is a set of sub-routes for a specified route. It can be used for inner
	// routes that share a common middleware or functionality that should be separate
	// from the parent gig instance while still inheriting from it.
	Group struct {
		common
		prefix     string
		middleware []MiddlewareFunc
		gig        *Gig
	}
)

// Use implements `Gig#Use()` for sub-routes within the Group.
func (g *Group) Use(middleware ...MiddlewareFunc) {
	g.middleware = append(g.middleware, middleware...)
	if len(g.middleware) == 0 {
		return
	}
	// Allow all requests to reach the group as they might get dropped if router
	// doesn't find a match, making none of the group middleware process.
	g.Handle("", NotFoundHandler)
	g.Handle("/*", NotFoundHandler)
}

// Handle implements `Gig#Handle()` for sub-routes within the Group.
func (g *Group) Handle(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.add(path, h, m...)
}

// Group creates a new sub-group with prefix and optional sub-group-level middleware.
func (g *Group) Group(prefix string, middleware ...MiddlewareFunc) *Group {
	m := make([]MiddlewareFunc, 0, len(g.middleware)+len(middleware))
	m = append(m, g.middleware...)
	m = append(m, middleware...)

	return g.gig.Group(g.prefix+prefix, m...)
}

// Static implements `Gig#Static()` for sub-routes within the Group.
func (g *Group) Static(prefix, root string) {
	g.static(prefix, root, g.Handle)
}

// File implements `Gig#File()` for sub-routes within the Group.
func (g *Group) File(path, file string) {
	g.file(path, file, g.Handle)
}

func (g *Group) add(path string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route {
	// Combine into a new slice to avoid accidentally passing the same slice for
	// multiple routes, which would lead to later add() calls overwriting the
	// middleware from earlier calls.
	m := make([]MiddlewareFunc, 0, len(g.middleware)+len(middleware))
	m = append(m, g.middleware...)
	m = append(m, middleware...)

	return g.gig.add(g.prefix+path, handler, m...)
}
