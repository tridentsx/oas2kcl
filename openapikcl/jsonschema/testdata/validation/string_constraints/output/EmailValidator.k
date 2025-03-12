import regex

schema EmailValidator:
    value: str

    check:
        regex.match(value, r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$") if value, "Value must be a valid email address"
