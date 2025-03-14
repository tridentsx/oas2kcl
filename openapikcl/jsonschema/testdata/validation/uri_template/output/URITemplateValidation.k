import regex

schema URITemplateValidation:
    # Validation of URI templates according to RFC 6570
    parameterizedTemplate?: str
    # Format: uri-template

    queryTemplate?: str
    # Format: uri-template

    simpleTemplate: str
    # Format: uri-template

    complexTemplate?: str
    # Format: uri-template

    multiLevelTemplate?: str
    # Format: uri-template


    check:
        simpleTemplate != None