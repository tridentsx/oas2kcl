import regex

schema Schema:
    website: URIValidator
    # Format: uri
    # URI string following RFC 3986


    check:
        website != None