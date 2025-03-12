import regex

schema DynamicConfig:
    # Configuration with dynamic properties following specific patterns
    name: str

    version: str

    check:
        # Regex pattern: ^\d+\.\d+\.\d+$
        name != None
        version != None
        version == None or regex.match("^\\d+\\.\\d+\\.\\d+$", version)

