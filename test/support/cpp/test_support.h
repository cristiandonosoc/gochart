#pragma once

#include <gtest/gtest.h>

#include <iostream>
#include <source_location>
#include <span>
#include <sstream>
#include <string>

namespace test
{

// Extracts the function name from a __FUNCTION__ like macro result.
// This is to make the identifier shorter.
std::string ExtractFunctionName(const std::string& function_name);

template <typename T>
std::string ToString(const T& t)
{
    std::stringstream ss;
    ss << t;
    return ss.str();
}

std::string ToString(const std::source_location& location);

template <typename T>
std::string PrintSpan(std::span<const T> things)
{
    std::stringstream ss;
    ss << "[" << std::endl;
    for (size_t i = 0; i < things.size(); i++)
    {
        const T& thing = things[i];
        ss << "\t" << ToString(thing) << std::endl;
    }
    ss << "]";

    return ss.str();
}

template <typename T>
std::string PrintContext(std::span<const T> want, std::span<const T> got, const std::source_location& location)
{
    std::stringstream ss;
    ss << std::endl
       << "Want: " << PrintSpan(want) << std::endl
       << "Got: " << PrintSpan(got) << std::endl
       << "Location: " << ToString(location);

    return ss.str();
}

template <typename T>
void CompareStates(std::span<const T> want,
                   std::span<const T> got,
                   std::source_location location = std::source_location::current())
{
    ASSERT_EQ(want.size(), got.size()) << PrintContext(want, got, location);

    for (size_t i = 0; i < want.size(); i++)
    {
        ASSERT_EQ(want[i], got[i]) << "Want: " << ToString(want[i]) << ", Got: " << ToString(got[i])
                                   << PrintContext(want, got, location);
    }
}

void CompareCalls(std::span<const std::string> want,
                  std::span<const std::string> got,
                  std::source_location location = std::source_location::current());

} // namespace test
