" *Very* simple syntax file, but gets a lot of the job done.

if exists("b:current_syntax")
	finish
endif

syntax keyword gochartStructural statechart trigger state transition capture
highlight link gochartStructural Type

syntax keyword gochartStatement enter exit transition capture
highlight link gochartStatement Statement

syntax keyword gochartConditional if do
highlight link gochartConditional Keyword

syntax match gochartComment "//.*$"
highlight link gochartComment Comment

let b:current_syntax = "gochart"

