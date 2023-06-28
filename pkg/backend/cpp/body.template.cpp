#include "header.template.h"

States Statechart::ParentState(States state) {
	// clang-format off
	switch (state) {
		{{- range.Statechart.States }}
		case {{.Name}}: return {{if .Parent}}{{.Parent}}{{else}}States::None{{end}};
		{{- end }}
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
