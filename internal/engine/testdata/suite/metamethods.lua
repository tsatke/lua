runTest("metamethods / add", function()
    local number1 = { 17 }
    local number2 = { 39 }
    local number_metatable = {
        ["__add"] = function(left, right)
            return left[1] + right[1]
        end
    }
    setmetatable(number1, number_metatable)

    assertEqual(56, number1 + number2)
    assertEqual(56, number2 + number1)
end)

runTest("metamethods / sub", function()
    local number1 = { 17 }
    local number2 = { 39 }
    local number_metatable = {
        ["__sub"] = function(left, right)
            return left[1] - right[1]
        end
    }
    setmetatable(number1, number_metatable)

    assertEqual(-22, number1 - number2)
    assertEqual(22, number2 - number1)
end)

runTest("metamethods / mul", function()
    local number1 = { 17 }
    local number2 = { 39 }
    local number_metatable = {
        ["__mul"] = function(left, right)
            return left[1] * right[1]
        end
    }
    setmetatable(number1, number_metatable)

    assertEqual(663, number1 * number2)
    assertEqual(663, number2 * number1)
end)

runTest("metamethods / div", function()
    local number1 = { 17 }
    local number2 = { 39 }
    local number_metatable = {
        ["__div"] = function(left, right)
            return left[1] / right[1]
        end
    }
    setmetatable(number1, number_metatable)

    assertEqual(0.4358974358974359, number1 / number2)
    assertEqual(2.2941176470588234, number2 / number1)
end)

runTest("metamethods / idiv", function()
    local number1 = { 17 }
    local number2 = { 39 }
    local number_metatable = {
        ["__idiv"] = function(left, right)
            return left[1] // right[1]
        end
    }
    setmetatable(number1, number_metatable)

    assertEqual(0, number1 // number2)
    assertEqual(2, number2 // number1)
end)

runTest("metamethods / mod", function()
    local number1 = { 17 }
    local number2 = { 39 }
    local number_metatable = {
        ["__mod"] = function(left, right)
            return left[1] % right[1]
        end
    }
    setmetatable(number1, number_metatable)

    assertEqual(17, number1 % number2)
    assertEqual(5, number2 % number1)
end)

runTest("metamethods / pow", function()
    local number1 = { 2 }
    local number2 = { 8 }
    local number_metatable = {
        ["__pow"] = function(left, right)
            return left[1] ^ right[1]
        end
    }
    setmetatable(number1, number_metatable)

    assertEqual(256, number1 ^ number2)
    assertEqual(64, number2 ^ number1)
end)

runTest("metamethods / bor", function()
    local number1 = { 2 }
    local number2 = { 3 }
    local number_metatable = {
        ["__bor"] = function(left, right)
            return left[1] | right[1]
        end
    }
    setmetatable(number1, number_metatable)

    assertEqual(3, number1 | number2)
    assertEqual(3, number2 | number1)
end)

runTest("metamethods / bxor", function()
    local number1 = { 2 }
    local number2 = { 3 }
    local number_metatable = {
        ["__bxor"] = function(left, right)
            return left[1] ~ right[1]
        end
    }
    setmetatable(number1, number_metatable)

    assertEqual(1, number1 ~ number2)
    assertEqual(1, number2 ~ number1)
end)

runTest("metamethods / band", function()
    local number1 = { 2 }
    local number2 = { 3 }
    local number_metatable = {
        ["__band"] = function(left, right)
            return left[1] & right[1]
        end
    }
    setmetatable(number1, number_metatable)

    assertEqual(2, number1 & number2)
    assertEqual(2, number2 & number1)
end)

runTest("metamethods / unm", function()
    local number1 = { 2 }
    local number_metatable = {
        ["__unm"] = function(val)
            return -val[1]
        end
    }
    setmetatable(number1, number_metatable)

    assertEqual(-2, -number1)
end)

runTest("metamethods / len", function()
    local value = { 1, 2, 3, "abc" }
    local value_metatable = {
        ["__len"] = function(val)
            return val[2]
        end
    }

    assertEqual(4, #value)

    setmetatable(value, value_metatable)

    assertEqual(2, #value)
    assertEqual(3, #value[4])
end)

runTest("metamethods / bnot", function()
    local number1 = { 2 }
    local number_metatable = {
        ["__bnot"] = function(val)
            return ~val[1]
        end
    }
    setmetatable(number1, number_metatable)

    assertEqual(-3, ~number1)
end)

runTest("metamethods / shl", function()
    local number1 = { 2 }
    local number2 = { 3 }
    local number_metatable = {
        ["__shl"] = function(left, right)
            return left[1] << right[1]
        end
    }
    setmetatable(number1, number_metatable)

    assertEqual(16, number1 << number2)
    assertEqual(12, number2 << number1)
end)

runTest("metamethods / shr", function()
    local number1 = { 1 }
    local number2 = { 3 }
    local number_metatable = {
        ["__shr"] = function(left, right)
            return left[1] >> right[1]
        end
    }
    setmetatable(number1, number_metatable)

    assertEqual(0, number1 >> number2)
    assertEqual(1, number2 >> number1)
end)

runTest("metamethods / concat", function()
    local string1 = { "a" }
    local string2 = { "b" }
    local string_metatable = {
        ["__concat"] = function(left, right)
            return left[1] .. right[1]
        end
    }
    setmetatable(string1, string_metatable)

    assertEqual("ab", string1 .. string2)
    assertEqual("ba", string2 .. string1)
end)

runTest("metamethods / eq", function()
    local string1 = { "a" }
    local string2 = { "b" }
    local string_metatable = {
        ["__eq"] = function()
            return true
        end
    }
    setmetatable(string1, string_metatable)

    assertTrue(string1 == string2)
    assertTrue(string2 == string1)

    local string3 = { "a" }
    local string4 = { "b" }
    string_metatable = {
        ["__eq"] = function()
            return false
        end
    }
    setmetatable(string3, string_metatable)

    assertTrue(string3 == string3) -- primitive equality takes precedence over metamethod
    assertFalse(string3 == string4)
end)

runTest("metamethods / lt", function()
    local number1 = { 17 }
    local number2 = { 39 }
    local number_metatable = {
        ["__lt"] = function(left, right)
            return left[1] < right[1]
        end
    }
    setmetatable(number1, number_metatable)

    assertTrue(number1 < number2)
    assertFalse(number2 < number1)
end)

runTest("metamethods / le (with lt)", function()
    local number1 = { 17 }
    local number2 = { 39 }
    local number_metatable = {
        ["__lt"] = function(left, right)
            return left[1] < right[1]
        end
    }
    setmetatable(number1, number_metatable)

    assertTrue(number1 <= number2)
    assertFalse(number2 <= number1)
end)

runTest("metamethods / le", function()
    local number1 = { 17 }
    local number2 = { 39 }
    local number_metatable = {
        ["__le"] = function(left, right)
            return left[1] <= right[1]
        end
    }
    setmetatable(number1, number_metatable)

    assertTrue(number1 <= number2)
    assertFalse(number2 <= number1)
end)

runTest("metamethods / newindex (function)", function()
    local tbl = {}
    local otherTbl = {}
    local tbl_metatable = {
        ["__newindex"] = function(t, key, val)
            otherTbl[key] = val
        end
    }
    setmetatable(tbl, tbl_metatable)

    tbl["foo"] = "bar"

    assertEqual(nil, tbl["foo"])
    assertEqual("bar", otherTbl["foo"])
end)

runTest("metamethods / newindex (table)", function()
    local tbl = {}
    local otherTbl = {}
    local tbl_metatable = {
        ["__newindex"] = otherTbl
    }
    setmetatable(tbl, tbl_metatable)

    tbl["foo"] = "bar"

    assertEqual(nil, tbl["foo"])
    assertEqual("bar", otherTbl["foo"])
end)

runTest("metamethods / index (function)", function()
    local tbl = {}
    local otherTbl = {
        ["foo"] = "bar"
    }
    local tbl_metatable = {
        ["__index"] = function(t, key)
            return otherTbl[key]
        end
    }
    setmetatable(tbl, tbl_metatable)

    assertEqual("bar", tbl["foo"])
    assertEqual("bar", otherTbl["foo"])
end)

runTest("metamethods / index (table)", function()
    local tbl = {}
    local otherTbl = {
        ["foo"] = "bar"
    }
    local tbl_metatable = {
        ["__index"] = otherTbl
    }
    setmetatable(tbl, tbl_metatable)

    assertEqual("bar", tbl["foo"])
    assertEqual("bar", otherTbl["foo"])
end)

runTest("metamethods / call", function()
    local tbl = {}
    local tbl_metatable = {
        ["__call"] = function(fn, arg1, arg2)
            return arg2, arg1
        end
    }
    setmetatable(tbl, tbl_metatable)

    r1, r2 = tbl("x", "y")
    assertEqual("y", r1)
    assertEqual("x", r2)
end)