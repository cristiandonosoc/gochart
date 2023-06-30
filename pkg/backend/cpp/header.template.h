// File generated by Gochart version "{{.Version}}" at {{.Time}}
// DO NOT MODIFY!

#pragma once

#include <array>
#include <assert.h>

namespace gochart {

class {{.ImplName}}
{
public:
    using ElementType = uint32_t;

public:
  enum class States {
		{{- range .Statechart.States}}
		{{.Name}},
		{{- end}}
		None,
	};

	{{- range .Statechart.Triggers }}
	struct Trigger{{.Name}} {
		{{- range .Args }}
		{{.Type}} {{.Name}};
		{{- end }}
	};
	{{- end }}

	static const char* ToString(States state);
	static States ParentState(States state);

private:
    class RingBuffer
    {
    public:
        bool IsEmpty() const { return ReadIndex == WriteIndex; }
        bool IsFull() const { return ((WriteIndex + 1) % TriggerQueue.size()) == ReadIndex; }

        void Enqueue(const ElementType &element);
        void Dequeue(ElementType *out);

    private:
        // TODO(cdc): Make queue size configurable.
        std::array<ElementType, 32> TriggerQueue;
        std::size_t ReadIndex;
        std::size_t WriteIndex;
    };

private:
};


} // namespace gochart
