import regex

schema URITemplateValidation:
    simpleTemplate: str
    complexTemplate?: str
    multiLevelTemplate?: str
    parameterizedTemplate?: str
    queryTemplate?: str

    check:
        regex.match(simpleTemplate, r"^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]") if simpleTemplate, "Value must be a valid URI"
        regex.match(complexTemplate, r"^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]") if complexTemplate, "Value must be a valid URI"
        regex.match(multiLevelTemplate, r"^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]") if multiLevelTemplate, "Value must be a valid URI"
        regex.match(parameterizedTemplate, r"^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]") if parameterizedTemplate, "Value must be a valid URI"
        regex.match(queryTemplate, r"^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]") if queryTemplate, "Value must be a valid URI"


