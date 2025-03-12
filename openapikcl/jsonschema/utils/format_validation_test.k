import datetime
import regex
import net

print("Testing all format validations:")

# Test date format (using datetime.validate)
print("\n-- Date Format Testing --")
valid_date = "2024-03-26"
invalid_date = "2024-13-26"
print("Valid date:", valid_date)
print("datetime.validate result:", datetime.validate(valid_date, "%Y-%m-%d"))
print("Invalid date:", invalid_date)
print("datetime.validate result:", datetime.validate(invalid_date, "%Y-%m-%d"))

# Test time format (using datetime.validate)
print("\n-- Time Format Testing --")
valid_time = "14:30:45"
invalid_time = "25:30:45"
print("Valid time:", valid_time)
print("datetime.validate result:", datetime.validate(valid_time, "%H:%M:%S"))
print("Invalid time:", invalid_time)
print("datetime.validate result:", datetime.validate(invalid_time, "%H:%M:%S"))

# Test date-time format (using regex)
print("\n-- Date-Time Format Testing --")
valid_datetime = "2024-03-26T14:30:45Z"
valid_datetime2 = "2024-03-26T14:30:45+02:00"
invalid_datetime = "2024-13-26T14:30:45Z"
datetime_pattern = r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$"
print("Valid date-time:", valid_datetime)
print("regex.match result:", regex.match(valid_datetime, datetime_pattern))
print("Valid date-time with timezone:", valid_datetime2)
print("regex.match result:", regex.match(valid_datetime2, datetime_pattern))
print("Invalid date-time:", invalid_datetime)
print("regex.match result:", regex.match(invalid_datetime, datetime_pattern))

# Test duration format (using regex)
print("\n-- Duration Format Testing --")
valid_duration = "P1Y2M3DT4H5M6S"
valid_duration2 = "PT4H5M"
invalid_duration = "P1X"
duration_pattern = r"^P(?:\d+(?:\.\d+)?Y)?(?:\d+(?:\.\d+)?M)?(?:\d+(?:\.\d+)?W)?(?:\d+(?:\.\d+)?D)?(?:T(?:\d+(?:\.\d+)?H)?(?:\d+(?:\.\d+)?M)?(?:\d+(?:\.\d+)?S)?)?$"
print("Valid duration:", valid_duration)
print("regex.match result:", regex.match(valid_duration, duration_pattern))
print("Valid duration (time only):", valid_duration2)
print("regex.match result:", regex.match(valid_duration2, duration_pattern))
print("Invalid duration:", invalid_duration)
print("regex.match result:", regex.match(invalid_duration, duration_pattern))

# Test email and IDN email (using regex)
print("\n-- Email Format Testing --")
valid_email = "user@example.com"
invalid_email = "user@example"
email_pattern = r"^[a-zA-Z0-9.!#$%&'*+/=?^_\{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
print("Valid email:", valid_email)
print("regex.match result:", regex.match(valid_email, email_pattern))
print("Invalid email:", invalid_email)
print("regex.match result:", regex.match(invalid_email, email_pattern))

# Test hostname and IDN hostname (using regex)
print("\n-- Hostname Format Testing --")
valid_hostname = "example.com"
valid_hostname2 = "sub.example.com"
invalid_hostname = "-example.com"
hostname_pattern = r"^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$"
print("Valid hostname:", valid_hostname)
print("regex.match result:", regex.match(valid_hostname, hostname_pattern))
print("Valid hostname with subdomain:", valid_hostname2)
print("regex.match result:", regex.match(valid_hostname2, hostname_pattern))
print("Invalid hostname:", invalid_hostname)
print("regex.match result:", regex.match(invalid_hostname, hostname_pattern))

# Test IPv4 and IPv6 (using net functions)
print("\n-- IP Address Format Testing --")
valid_ipv4 = "192.168.1.1"
invalid_ipv4 = "192.168.1.300"
valid_ipv6 = "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
print("Valid IPv4:", valid_ipv4)
print("net.is_IPv4 result:", net.is_IPv4(valid_ipv4))
print("Invalid IPv4:", invalid_ipv4)
print("net.is_IPv4 result:", net.is_IPv4(invalid_ipv4))
print("Valid IPv6:", valid_ipv6)
print("net.is_IP result:", net.is_IP(valid_ipv6))

# Test URI and URI reference (using regex)
print("\n-- URI Format Testing --")
valid_uri = "https://example.com/path?query=value#fragment"
valid_uri_reference = "/path?query=value#fragment"
invalid_uri = "://example.com"
uri_pattern = r"^[a-zA-Z][a-zA-Z0-9+.-]*:[^\s]*$"
uri_reference_pattern = r"^(?:[a-zA-Z][a-zA-Z0-9+.-]*:)?(?://[^\s/$.?#].[^\s]*|[^\s/$.?#].[^\s]*)$"
print("Valid URI:", valid_uri)
print("regex.match result:", regex.match(valid_uri, uri_pattern))
print("Valid URI reference:", valid_uri_reference)
print("regex.match result:", regex.match(valid_uri_reference, uri_reference_pattern))
print("Invalid URI:", invalid_uri)
print("regex.match result:", regex.match(invalid_uri, uri_pattern))

# Test UUID (using regex)
print("\n-- UUID Format Testing --")
valid_uuid = "f47ac10b-58cc-4372-a567-0e02b2c3d479"
invalid_uuid = "f47ac10b-58cc-4372-a567-0e02b2c3d47"
uuid_pattern = r"^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"
print("Valid UUID:", valid_uuid)
print("regex.match result:", regex.match(valid_uuid, uuid_pattern))
print("Invalid UUID:", invalid_uuid)
print("regex.match result:", regex.match(invalid_uuid, uuid_pattern))

# Test JSON Pointer and Relative JSON Pointer (using regex)
print("\n-- JSON Pointer Format Testing --")
valid_json_pointer = "/foo/bar/0"
invalid_json_pointer = "foo/bar"
valid_relative_json_pointer = "2/foo/bar"
json_pointer_pattern = r"^(?:/(?:[^~/]|~0|~1)*)*$"
relative_json_pointer_pattern = r"^(?:0|[1-9][0-9]*)(?:/(?:[^~/]|~0|~1)*)*$"
print("Valid JSON Pointer:", valid_json_pointer)
print("regex.match result:", regex.match(valid_json_pointer, json_pointer_pattern))
print("Invalid JSON Pointer:", invalid_json_pointer)
print("regex.match result:", regex.match(invalid_json_pointer, json_pointer_pattern))
print("Valid Relative JSON Pointer:", valid_relative_json_pointer)
print("regex.match result:", regex.match(valid_relative_json_pointer, relative_json_pointer_pattern))

print("\nAll format validations tested successfully!") 