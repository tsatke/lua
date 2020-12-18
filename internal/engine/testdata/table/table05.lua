a = {}
function get_a()
    return a
end

get_a().b = "foobar"
print(get_a().b)