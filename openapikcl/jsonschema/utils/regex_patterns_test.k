import regex

# Test RFC 3339 date-time format
datetime_pattern = r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$"
print("DateTime pattern:", datetime_pattern)

# Test match result
match_result = regex.match("2024-03-20T10:00:00Z", datetime_pattern)
print("Match result for '2024-03-20T10:00:00Z':", match_result)

# Try escaping the pattern
datetime_pattern_escaped = r"^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}(\\.\\d+)?(Z|[+-]\\d{2}:\\d{2})$"
print("Escaped pattern:", datetime_pattern_escaped)
match_result_escaped = regex.match(datetime_pattern_escaped, "2024-03-20T10:00:00Z")
print("Match result with escaped pattern:", match_result_escaped)

# Try simpler pattern
simple_pattern = r"^\d{4}-\d{2}-\d{2}$"
print("Simple date pattern:", simple_pattern)
match_result_simple = regex.match("2024-03-20", simple_pattern)
print("Match result for simple pattern:", match_result_simple)

# Valid date-time values
assert regex.match("2024-03-20T10:00:00Z", datetime_pattern), "Valid UTC date-time should match"
assert regex.match("2024-03-20T10:00:00+01:00", datetime_pattern), "Valid timezone offset date-time should match"
assert regex.match("2024-03-20T10:00:00.123Z", datetime_pattern), "Valid date-time with fractional seconds should match"
assert regex.match("2024-03-20T10:00:00.123+01:00", datetime_pattern), "Valid date-time with fractional seconds and timezone should match"

# Invalid date-time values
assert regex.match("2024-13-20T10:00:00Z", datetime_pattern) == False
assert not regex.match("2024-03-20T25:00:00Z", datetime_pattern), "Invalid hour should not match"
assert not regex.match("2024-03-20T10:00:00", datetime_pattern), "Missing timezone should not match"
assert not regex.match("2024-03-20", datetime_pattern), "Missing time should not match"

# Test IPv4 format
ipv4_pattern = r"^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$"
print("IPv4 pattern:", ipv4_pattern)
match_result_ipv4 = regex.match("192.168.1.1", ipv4_pattern)
print("Match result for '192.168.1.1':", match_result_ipv4)

# Valid IPv4 values
assert regex.match("192.168.1.1", ipv4_pattern), "Valid IPv4 should match"
assert regex.match("10.0.0.0", ipv4_pattern), "Valid IPv4 should match"
assert regex.match("172.16.254.1", ipv4_pattern), "Valid IPv4 should match"
assert regex.match("0.0.0.0", ipv4_pattern), "Valid IPv4 should match"
assert regex.match("255.255.255.255", ipv4_pattern), "Valid IPv4 should match"

# Invalid IPv4 values
assert not regex.match("256.256.256.256", ipv4_pattern), "Invalid octets should not match"
assert not regex.match("1.2.3", ipv4_pattern), "Missing octet should not match"
assert not regex.match("192.168.001.1", ipv4_pattern), "Leading zeros should not match"
assert not regex.match("192.168.1.1.1", ipv4_pattern), "Extra octet should not match"

# Test email format
email_pattern = r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"
print("Email pattern:", email_pattern)
match_result_email = regex.match("test@example.com", email_pattern)
print("Match result for 'test@example.com':", match_result_email)

# Valid email values
assert regex.match("test@example.com", email_pattern), "Valid email should match"
assert regex.match("user.name@domain.co.uk", email_pattern), "Valid email with dots should match"
assert regex.match("user+tag@example.com", email_pattern), "Valid email with plus should match"
assert regex.match("user@sub.domain.com", email_pattern), "Valid email with subdomain should match"

# Invalid email values
assert not regex.match("invalid", email_pattern), "Not an email should not match"
assert not regex.match("@example.com", email_pattern), "Missing local part should not match"
assert not regex.match("user@", email_pattern), "Missing domain should not match"
assert not regex.match("user@.com", email_pattern), "Invalid domain should not match"

# Test URI format
uri_pattern = r"^(https?|ftp)://[^\s/$.?#].[^\s]*$"
print("URI pattern:", uri_pattern)
match_result_uri = regex.match("http://example.com", uri_pattern)
print("Match result for 'http://example.com':", match_result_uri)

# Valid URI values
assert regex.match("http://example.com", uri_pattern), "Valid http URI should match"
assert regex.match("https://example.com", uri_pattern), "Valid https URI should match"
assert regex.match("ftp://example.com", uri_pattern), "Valid ftp URI should match"
assert not regex.match("mailto:user@example.com", uri_pattern), "mailto URI should not match this pattern"

# Invalid URI values
assert not regex.match("example.com", uri_pattern), "Missing scheme should not match"
assert not regex.match("http:/example.com", uri_pattern), "Invalid format should not match"
assert not regex.match("http//example.com", uri_pattern), "Missing colon should not match"
assert not regex.match("http:example.com", uri_pattern), "Missing slashes should not match"

# Test UUID format
uuid_pattern = r"^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
print("UUID pattern:", uuid_pattern)
match_result_uuid = regex.match("123e4567-e89b-12d3-a456-426614174000", uuid_pattern)
print("Match result for '123e4567-e89b-12d3-a456-426614174000':", match_result_uuid)

# Valid UUID values
assert regex.match("123e4567-e89b-12d3-a456-426614174000", uuid_pattern), "Valid UUID should match"
assert regex.match("550e8400-e29b-41d4-a716-446655440000", uuid_pattern), "Valid UUID should match"
assert regex.match("6ba7b810-9dad-11d1-80b4-00c04fd430c8", uuid_pattern), "Valid UUID should match"

# Invalid UUID values
assert not regex.match("not-a-uuid", uuid_pattern), "Invalid format should not match"
assert not regex.match("123e4567-e89b-12d3-a456-42661417400", uuid_pattern), "Too short should not match"
assert not regex.match("123e4567-e89b-12d3-a456-4266141740000", uuid_pattern), "Too long should not match"
assert not regex.match("123e4567-e89b-12d3-a456-42661417400g", uuid_pattern), "Invalid character should not match"

print("All regex pattern tests passed!")

print("Testing regex patterns for format validation:")

# Test simple pattern
pattern1 = "^test$"
test_str1 = "test"
test_str2 = "test123"
print("\nPattern:", pattern1)
print("String:", test_str1)
print("Match:", regex.match(test_str1, pattern1))
print("String:", test_str2)
print("Match:", regex.match(test_str2, pattern1))

# Test date pattern
pattern2 = "^2024-03-20$"
test_str3 = "2024-03-20"
print("\nPattern:", pattern2)
print("String:", test_str3)
print("Match:", regex.match(test_str3, pattern2))

# Test properly escaped digit pattern
pattern3 = "^\\d{4}-\\d{2}-\\d{2}$"
test_str4 = "2024-03-20"
print("\nEscaped pattern:", pattern3)
print("String:", test_str4)
print("Match:", regex.match(test_str4, pattern3))

# Test unescaped digit pattern with r prefix
pattern4 = r"^\d{4}-\d{2}-\d{2}$"
print("\nUnescaped pattern with r prefix:", pattern4)
print("String:", test_str4)
print("Match:", regex.match(test_str4, pattern4))

print("\nAll tests completed successfully.") 
