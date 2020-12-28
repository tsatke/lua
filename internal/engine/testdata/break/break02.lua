do
    for i, v in ipairs({ 1, 2, 3, 4, 5, 6, 7 }) do
        print(i, v)
        if v > 2 then
            break
        end
    end
    print("end")
end