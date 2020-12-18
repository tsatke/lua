function a()
    if true then
        return "hello"
    end
    error("this must not happen")
end

return a()