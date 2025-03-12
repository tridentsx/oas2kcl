import regex

schema SchemaWebsite:
    value: str

    check:
        regex.match(value, r"^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]") if value, "Value must be a valid URI"
