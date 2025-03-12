

import regex


# Generated schemas for format validation
# Email validator schema
schema EmailValidator:
    value: str

    check:
        regex.match(value, r"^[a-zA-Z0-9.!#$%&'*+/=?^_'{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"), "Value must be a valid email address"


# Validator schema for Schema_emailsValidator
schema Schema_emailsValidator:
    value?: [EmailValidator] = []

    check:
        # Validate email format for all array items
        all item in value { regex.match(item, r"^[a-zA-Z0-9.!#$%&'*+/=?^_'{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$") } if value, "All items must be valid email addresses"
