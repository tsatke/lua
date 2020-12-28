do
    local i = 0
    repeat
        i = i + 1
        print(i)

        if i > 3 then
            break
        end
    until i == 7
    print("end")
end