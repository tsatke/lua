-- This file contains utility functions for tests.

--- Checks equality between left and right. If not equal, this
--- will error.
function assertEqual(left, right)
    if left ~= right then
        error("not equal: want " .. left .. " but got " .. right, 2)
    end
end

assertEqual(1, 1)
assertEqual("abc", "abc")
if pcall(assertEqual, 1, 2) then
    error("EQ seems broken")
end

--- Checks the argument to be true. If the argument is not
--- equal to true, this will error.
function assertTrue(exp)
    if exp ~= true then
        error("not true", 2)
    end
end

assertTrue(true)
if pcall(assertTrue, false) then
    error("assertTrue seems broken")
end

--- Checks the argument to be false. If the argument is not
--- equal to false, this will error.
function assertFalse(exp)
    if exp ~= false then
        error("not false", 2)
    end
end

assertFalse(false)
if pcall(assertFalse, true) then
    error("assertFalse seems broken")
end

function runTest(name, testFn)
    print("runTest [" .. name .. "]")
    ok, msg = pcall(testFn)
    if not ok then
        print("❌ FAIL (message: " .. msg .. ")")
        _RC = 1
    else
        print("✅ OK")
    end
end