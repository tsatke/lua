-- This file contains utility functions for tests.

function EQ(left, right)
    if left ~= right then
        error("not equal", 2)
    end
end

-- verify internal consistency of EQ

EQ(1, 1)
EQ("abc", "abc")
if pcall(EQ, 1, 2) then
    error("EQ seems broken")
end