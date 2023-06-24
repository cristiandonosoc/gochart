// Package gochart_lang is the frontend for what we can "gochart_lang", which is a custom spec
// language to describe statecharts.
package gochart_lang

import ()

type GochartLangFrontend struct {
}

// NewFrontend returns a frontend capable of parting the gochart language.
func NewGochartLangFrontend() *GochartLangFrontend {
	return &GochartLangFrontend{}
}
