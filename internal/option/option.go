// Package option is a generic definition of option that can be used by frontend and backend to
// export their options
package option

type Option func(opts interface{})
