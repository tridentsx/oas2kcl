import regex

schema DynamicConfig:
    # Configuration with dynamic properties following specific patterns
    name: str

    version: str
    # Regex pattern: ^\d+\.\d+\.\d+$


    check:
        name != None
        version != None
        version == None or regex.match("^\\d+\\.\\d+\\.\\d+$", version)