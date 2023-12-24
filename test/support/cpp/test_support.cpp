#include "test_support.h"

#include <iomanip>

namespace test
{

namespace
{

std::string SideBySidePrint(std::span<const std::string> want, std::span<const std::string> got)
{
    constexpr size_t kWidth = 50;


    std::stringstream ss;
    ss << std::left << std::setw(kWidth) << "WANT" << std::setw(0) << " | " << std::setw(kWidth) << "GOT" << std::endl;

    ss << std::setw(0);

    size_t max = std::max(want.size(), got.size());

    for (size_t i = 0; i < max; i++)
    {
        std::string want_item = i < want.size() ? want[i] : "<empty>";
        std::string got_item = i < got.size() ? got[i] : "<empty>";


        ss << std::setw(kWidth) << want_item << std::setw(0) << " | " << std::setw(kWidth) << got_item;
        ss << std::setw(0);

        if (want_item != got_item)
        {
            ss << " <--- DIFF!";
        }

        ss << std::endl;
    }

    return ss.str();
}

} // namespace

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
    ASSERT_EQ(want.size(), got.size()) << SideBySidePrint(want, got);

    for (size_t i = 0; i < want.size(); i++)
    {
        ASSERT_EQ(want[i], got[i]) << "Want: " << ToString(want[i]) << ", Got: " << ToString(got[i])
                                   << PrintContext(want, got, location);
    }
}

} // namespace test
