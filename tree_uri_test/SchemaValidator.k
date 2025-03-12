

import regex


# Generated schemas for format validation
# URI validator schema
schema URIValidator:
    value: str

    check:
        # URI format validation - checks for valid URI format with scheme
        regex.match(value, r"^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]"), "Value must be a valid URI"


# Validator schema for SchemaValidator
schema SchemaValidator:
    website: URIValidator

    check:
    check:
        # Validate URI format for website
        regex.match(website, r"^(https?|ftp|file)://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%!=(string=website)~_|]") if website, "%!s(MISSING) must be a valid URI"
