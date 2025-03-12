import regex

schema URIValidator:
    value: str

    check:
        # URI format validation - checks for valid URI format with scheme
        regex.match(value, r"^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]") if value, "Value must be a valid URI"
