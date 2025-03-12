

import regex


# Generated schemas for format validation
# Email validator schema
schema EmailValidator:
    value: str

    check:
        regex.match(value, r"^[a-zA-Z0-9.!#$%&'*+/=?^_'{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"), "Value must be a valid email address"


# Validator schema for Schema_emails_itemsValidator
schema Schema_emails_itemsValidator:
    value: str

    check:
        # Validate email format
        regex.match(value, r"^[a-zA-Z0-9.!#$%&'*+/=?^_'{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$") if value, "Value must be a valid email address"
