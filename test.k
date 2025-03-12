schema TestSchema:
    value: str

    check:
        value == None or len(value) >= 3, "value must have a minimum length of 3"

test = TestSchema {
    value = "abc"
}