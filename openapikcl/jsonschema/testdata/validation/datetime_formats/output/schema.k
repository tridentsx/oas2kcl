import datetime

import regex

schema DateTimeValidator:
    value?: str

    check:
        regex.match(value, r"^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])T([01]\d|2[0-3]):[0-5]\d:[0-5]\d(\.\d+)?(Z|[+-]([01]\d|2[0-3]):[0-5]\d)$") if value, "Value must be a valid RFC 3339 date-time"
        value != None and value and datetime.validate(value[:10], "%Y-%m-%d"), "Value contains an invalid date component"


