if exists("b:current_syntax")
	finish
endif

syntax keyword gochartStructural gochart nextgroup=gochartStructuralIdentifier
syntax keyword gochartStructural state nextgroup=gochartStructuralIdentifier

" Structural Options
syntax match gochartStructuralIdentifier "\s\+\k\+" contained nextgroup=gochartStructuralOptionsRegion
syntax region gochartStructuralOptionsRegion start="(" end=")" contained contains=gochartStructuralOptionKey
syntax match gochartStructuralOptionKey "\k\+" contained nextgroup=gochartStructuralOptionAssignment
syntax match gochartStructuralOptionAssignment "\s*=\s*" contained nextgroup=gochartStructuralOptionValue
syntax match gochartStructuralOptionValue "\k\+" contained

" Trigger
syntax keyword gochartStructural trigger nextgroup=gochartTriggerIdentifier
syntax match gochartTriggerIdentifier "\s\+\k\+" contained nextgroup=gochartTriggerArgsRegion
syntax region gochartTriggerArgsRegion start="(" end=")" keepend contained contains=gochartTriggerArgs
syntax match gochartTriggerArgs ".*" contained

" Transition
syntax keyword gochartStructural transition nextgroup=gochartTransitionNullTrigger,gochartTransitionTrigger
syntax match gochartTransitionTrigger "\s\+\k\+" contained nextgroup=gochartTransitionArrow
syntax match gochartTransitionNullTrigger "\s\+\<null\>" contained nextgroup=gochartTransitionArrow
syntax match gochartTransitionArrow "\s*>\s*" contained nextgroup=gochartTransitionState
syntax match gochartTransitionState "\k\+" contained nextgroup=gochartTransitionIf,gochartTransitionDo
syntax match gochartTransitionIf "\s\+\<if\>\s\+" contained nextgroup=gochartTransitionIfCodeRegion
syntax region gochartTransitionIfCodeRegion start="{" end="}" keepend contained contains=gochartTransitionIfCode nextgroup=gochartTransitionDo
syntax match gochartTransitionIfCode ".*" contained
syntax match gochartTransitionDo "\s\+\<do\>\s\+" contained nextgroup=gochartTransitionDoCodeRegion
syntax region gochartTransitionDoCodeRegion start="{" end="}" keepend contained contains=gochartTransitionDoCode
syntax match gochartTransitionDoCode ".*" contained

syntax keyword gochartStatement enter exit capture

highlight link gochartStructural Statement
highlight link gochartStructuralIdentifier Identifier
highlight link gochartStructuralOptionKey Special
highlight link gochartStructuralOptionValue Constant

highlight link gochartTriggerIdentifier Identifier
highlight link gochartTriggerArgs Float

highlight link gochartTransitionNullTrigger Keyword
highlight link gochartTransitionTrigger Identifier
highlight link gochartTransitionState Identifier
highlight link gochartTransitionIf Statement
highlight link gochartTransitionIfCode Float
highlight link gochartTransitionDo Statement
highlight link gochartTransitionDoCode Float

highlight link gochartStatement Statement

syntax region gochartCommand start="//" end="$" contains=@Spell
highlight link gochartComment Comment

let b:current_syntax = "gochart"

