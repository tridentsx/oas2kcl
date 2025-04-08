import regex

schema DynamicConfig:
    version: str
    name: str

    check:
        regex.match(version, r"^\d+\.\d+\.\d+$") if version, "String must match pattern ^\d+\.\d+\.\d+$"

