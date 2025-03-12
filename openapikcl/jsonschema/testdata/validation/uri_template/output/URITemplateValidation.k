import regex

schema URITemplateValidation:
    # Validation of URI templates according to RFC 6570
    complexTemplate?: str
    # Format: uri-template
    multiLevelTemplate?: str
    # Format: uri-template
    parameterizedTemplate?: str
    # Format: uri-template
    queryTemplate?: str
    # Format: uri-template
    simpleTemplate: str

    check:
        # Format: uri-template
        simpleTemplate != None

