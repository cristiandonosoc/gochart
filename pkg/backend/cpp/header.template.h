// File generated by Gochart version "{{.Version}}" at {{.Time}}
// DO NOT MODIFY!

#pragma once

#include <array>
#include <assert.h>

class Statechart
{
public:
    using ElementType = uint32_t;

public:
  // clang-format off
  enum class States {
		None,
		{{- range.Statechart.States}}
		{{.Name}},
		{{- end}}
	};
  // clang-format on

private:
	States ParentState(States state);

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
