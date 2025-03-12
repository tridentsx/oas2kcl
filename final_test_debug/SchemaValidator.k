

import regex


# Generated schemas for format validation
# Email validator schema
schema EmailValidator:
    value: str

    check:
        regex.match(value, r"^[a-zA-Z0-9.!#$%&'*+/=?^_'{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"), "Value must be a valid email address"


# Validator schema for SchemaValidator
schema SchemaValidator:
    emails: [EmailValidator] = []

    check:
        all item in emails { regex.match(`^[a-zA-Z0-9.!#$%&'*+/=?^_'{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`, item) } if emails, "All items in emails must be valid email addresses"
    check:
        all item in emails { regex.match(`^[a-zA-Z0-9.!#$%&'*+/=?^_'{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`, item) } if emails, "All items in emails must be valid email addresses"
