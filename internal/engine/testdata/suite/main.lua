-- This is the main entry point to the Lua test
-- suite. The engine uses these files to run
-- something that comes close to internal consistency
-- tests.
-- Lua code is executed, and will error out if something
-- is wrong. The engine will only fail the test if the
-- lua code errors. Otherwise, it will always pass,
-- independently of output to stdout and stderr.

function runTests()
    print("no functional code, this is only a skeleton")
end

runTests()