// File generated by Gochart version "{{.Version}}" at {{.Time}}
// DO NOT MODIFY!

#include "{{.HeaderInclude}}"

#include <cassert>

// TODO(cdc): This is very simple, but something fancier to support more compilers could be needed.
#ifdef _MSC_VER
#define DEBUG_BREAK __debugbreak()
#else
#define DEBUG_BREAK ___builtin_debugtrap()
#endif

Statechart::States Statechart::ParentState(States state)
{
    // clang-format off
	switch (state) {
		{{- range.Statechart.States }}
		case States::{{.Name}}: return {{if .Parent}}States::{{.Parent.Name}}{{else}}States::None{{end}};
		{{- end }}
		case States::None: DEBUG_BREAK; return States::None;
	}
    // clang-format on
}

// Statechart::RingBuffer --------------------------------------------------------------------------

void Statechart::RingBuffer::Enqueue(const ElementType &element)
{
    assert(!IsFull());
    TriggerQueue[WriteIndex] = element;
    WriteIndex = (WriteIndex + 1) % TriggerQueue.size();
}

void Statechart::RingBuffer::Dequeue(ElementType *out)
{
    assert(!IsEmpty());

    *out = TriggerQueue[ReadIndex];
    ReadIndex = (ReadIndex + 1) % TriggerQueue.size();
}
