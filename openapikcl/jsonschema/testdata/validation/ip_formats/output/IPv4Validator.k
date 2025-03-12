import regex

schema IPv4Validator:
    value: str

    check:
        # Validate IPv4 format
        regex.match(value, r"^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$") if value, "Value must be a valid IPv4 address"
