#include "delegate.h"

#include "concurrent.define.h"

Delegate::Delegate()
{
}

Delegate::~Delegate() = default;

void Delegate::Begin()
{
    Gochart->Activate();
}

void Delegate::End()
{
    Gochart->Deactivate();
}

