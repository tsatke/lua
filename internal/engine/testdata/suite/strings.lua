runTest("strings", function()
    assertEqual("ab", "a" .. "b")
    assertEqual(7, #"abcdefg")
end)