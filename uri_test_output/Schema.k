import regex

schema Schema:
    website: str
    # uri validation for website

    check:
        website == None or regex.match("^(https?|ftp)://[^\s/$.?#].[^\s]*$", website)

        website != None