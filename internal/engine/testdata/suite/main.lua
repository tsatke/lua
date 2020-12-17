-- This is the main entry point to the Lua test
-- suite. The engine uses these files to run
-- something that comes close to internal consistency
-- tests.
-- Lua code is executed, and will error out if something
-- is wrong. The engine will only fail the test if the
-- lua code errors. Otherwise, it will always pass,
-- independently of output to stdout and stderr.

do
    print("verifying integrity of core functions")
    if pcall(function()
        error("error to make pcall return false")
    end) then
        error("pcall seems broken")
    end

    if false then
        error("if seems broken")
    end
    if 4 == 5 then
        error("== seems broken")
    end
    if not 3 == 3 then
        error("not seems broken")
    end

    local x = false
    local ret = pcall(function()
        error("some error")
        x = true
    end)
    if x then
        if not ret then
            error("error seems broken")
        else
            print("error, pcall or both seem broken")
            os.exit(1)
        end
    end
end

print("loading utilities")
dofile("utility.lua")

print("starting suite")

dofile("operators.lua")