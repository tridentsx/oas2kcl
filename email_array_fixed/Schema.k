import regex

schema Schema:
    emails: [EmailValidator]
    # Item format: email


    check:
        emails != None