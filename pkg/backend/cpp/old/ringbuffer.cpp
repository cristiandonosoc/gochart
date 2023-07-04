// {{.ImplName}}::RingBuffer --------------------------------------------------------------------------

void {{.ImplName}}::RingBuffer::Enqueue(const ElementType &element)
{
    assert(!IsFull());
    TriggerQueue[WriteIndex] = element;
    WriteIndex = (WriteIndex + 1) % TriggerQueue.size();
}

void {{.ImplName}}::RingBuffer::Dequeue(ElementType *out)
{
    assert(!IsEmpty());

    *out = TriggerQueue[ReadIndex];
    ReadIndex = (ReadIndex + 1) % TriggerQueue.size();
}


