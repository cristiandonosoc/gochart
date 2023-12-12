#include "delegate.h"

#include <gtest/gtest.h>

TEST(ConcurrentTest, Hello)
{
    Delegate delegate;
    delegate.Begin();
    delegate.End();
}
