# No schema imports needed - schemas in same directory

schema Schema:
    permissions?: any
    concurrency?: any
    defaults?: any
    env?: any
    jobs: {str:any}
    name?: str
    on: any

    check:
        # oneOf validation - exactly one of these schemas must match
        # Field 'concurrency' must conform to exactly one of 2 schema options
        # Option 1 schema:
        # Option 2 schema:
        # KCL doesn't directly support 'exactly one of' validation without a discriminator.
        # Consider adding custom validation logic for this field.
        len(jobs) >= 1
        # oneOf validation - exactly one of these schemas must match
        # Field 'on' must conform to exactly one of 3 schema options
        # Option 1 schema:
        # Option 2 schema:
        # Option 3 schema:
        # KCL doesn't directly support 'exactly one of' validation without a discriminator.
        # Consider adding custom validation logic for this field.