# KCL Validation Patterns

This document contains common validation patterns in KCL, based on community knowledge and best practices.

## List Validation

### List Definition and Type
```kcl
# List of strings
[str]

# List of custom types
[Item]  # where Item is a schema
```

### List Length Validation
```kcl
check:
    len(items) > 3, "Must have more than 3 items"
    len(items) < 10, "Must have less than 10 items"
```

### List Uniqueness Check
```kcl
check:
    isunique(items), "Items must be unique"
```

### Validating Each Item in a List
```kcl
check:
    all myitem in items { 
        condition 
    }, "error message"
```

### Examples of List Item Validation
```kcl
# String length check
all myitem in items { 
    len(myitem.name) > 2 
}, "Invalid name: too short"

# Pattern matching
all myitem in items { 
    regex.match(myitem.name, "[a-z]+") 
}, "Invalid name: can only use lowercase letters"

# Numeric range check
all myitems in items { 
    0 <= myitems.value <= 100 
}
```

### Complete Example Combining Multiple Validations
```kcl
schema Item:
    name: str
    value: int

schema Config:
    items: [Item]

    check:
        isunique(items), "Items must be unique"
        len(items) > 3, "Must have more than 3 items"
        len(items) < 10, "Must have less than 10 items"
        all myitem in items { 
            len(myitem.name) > 2 
        }, "Invalid name: too short"
        all myitem in items { 
            regex.match(myitem.name, "[a-z]+") 
        }, "Invalid name: can only use lowercase letters"
        all myitems in items { 
            0 <= myitems.value <= 100 
        }
```

## Extended Primitives

For extended primitives (like DateTime, Email, URL, etc.), define a dedicated schema:

```kcl
import regex

schema DateTime:
    value: str

    check:
        regex.match(r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$", value) == true, 
            "Invalid datetime format: Must follow ISO 8601 (e.g., '2024-03-10T15:30:00Z')"

# Usage in main schema
schema MainSchema:
    dates: [DateTime]  # List of validated primitives
```

### Key Points for Extended Primitives
- Define a dedicated schema for each extended primitive
- Include validation rules in the schema's `check` block
- Use regex patterns for format validation
- These schemas can be in the same file or a separate file in the same folder (no explicit import needed)
- Each extended primitive gets its own schema
- Validation happens at the schema level
- Error messages should be clear and descriptive
- These schemas can be reused across different parts of the codebase

## Required vs Optional Properties

Properties with default values are mandatory and cannot be None:
```kcl
schema TestSchema:
    name: str = "default"     # mandatory
    age?: int                 # optional

    check:
        0 <= age if age       # only check if age is provided
        name != ""            # check non-empty string
        name is not None      # redundant since name has default value
``` 