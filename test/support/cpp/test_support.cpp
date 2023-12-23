#include "test_support.h"

namespace test
{

std::string ExtractFunctionName(const std::string& function_name)
{
    size_t pos = function_name.rfind("::");
    if (pos != std::string::npos)
    {
        // +2 to skip "::".
        return function_name.substr(pos + 2);
    }
    // Return the original name if "::" not found.
    return function_name;
}

std::string ToString(const std::source_location& location)
{
    std::stringstream ss;
    ss << location.file_name() << ":" << location.line();
    return ss.str();
}

void CompareCalls(std::span<const std::string> want,
                  std::span<const std::string> got,
                  std::source_location location)
{
    ASSERT_EQ(want.size(), got.size()) << PrintContext(want, got, location);

    for (size_t i = 0; i < want.size(); i++)
    {
        ASSERT_EQ(want[i], got[i]) << "Want: " << ToString(want[i]) << ", Got: " << ToString(got[i])
                                   << PrintContext(want, got, location);
    }
}

} // namespace test
