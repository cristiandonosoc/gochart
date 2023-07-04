#include <array>

class RingBuffer
{
    using ElementType = uint32_t;
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
