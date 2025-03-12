

import regex


# Generated schemas for format validation
# URI validator schema
schema URIValidator:
    value: str

    check:
        # URI format validation - checks for valid URI format with scheme
        regex.match(value, r"^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]"), "Value must be a valid URI"


# Validator schema for Schema_websiteValidator
schema Schema_websiteValidator:
    value: str

    check:
        # Validate URI format
        regex.match(value, r"^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]") if value, "Value must be a valid URI"
