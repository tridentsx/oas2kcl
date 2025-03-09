import regex

schema CompositionTest:
    """Schema to test various composition features."""
    base: str  # Base property present in all variants
    version: str
    legacy?: bool
    
    # Configuration fields (anyOf)
    type?: str
    configA?: {
        feature: str
    }
    configB?: [str]
    
    # Mode fields (oneOf)
    mode?: str
    simpleValue?: str
    advancedConfig?: {
        setting1: bool
        setting2: float
    }
    
    check:
        # Base validation
        regex.match(version, r"^v\d+\.\d+\.\d+$")
        
        # allOf conditional validation
        if regex.match(version, r"^v1\."):
            legacy exists
            legacy == True
        
        # anyOf validation - at least one configuration must be valid
        type exists
        if type == "A":
            configA exists
            if configA:
                configA.feature exists
        elif type == "B":
            configB exists
            if configB:
                len(configB) >= 1
        
        # oneOf validation - exactly one mode must be used
        mode exists
        mode in ["simple", "advanced"]
        
        if mode == "simple":
            simpleValue exists
            advancedConfig not exists
            
        if mode == "advanced":
            advancedConfig exists
            simpleValue not exists
            
            if advancedConfig:
                advancedConfig.setting1 exists
                advancedConfig.setting2 exists