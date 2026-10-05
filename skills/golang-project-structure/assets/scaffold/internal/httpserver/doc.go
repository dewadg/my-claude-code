// Package httpserver owns the HTTP transport: the router, the name-tagged
// route groups domains mount onto, global middleware and the server
// lifecycle. Domains never receive the router itself.
package httpserver
