import regex

schema Schema:
    emails: list
    # Each item should be a valid email format
    # KCL would need custom validation for this format

    check:
        emails != None