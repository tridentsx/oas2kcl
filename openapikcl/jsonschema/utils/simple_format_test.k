import datetime
import regex
import net

print("Simple Format Validation Test")

# Test date format
print("\nDate Validation:")
valid_date = "2024-03-26"
print("Valid date:", valid_date)
print("datetime.validate result:", datetime.validate(valid_date, "%Y-%m-%d"))

# Test time format
print("\nTime Validation:")
valid_time = "14:30:45"
print("Valid time:", valid_time)
print("datetime.validate result:", datetime.validate(valid_time, "%H:%M:%S"))

# Test IPv4 validation
print("\nIPv4 Validation:")
valid_ipv4 = "192.168.1.1"
print("Valid IPv4:", valid_ipv4)
print("net.is_IPv4 result:", net.is_IPv4(valid_ipv4))

# Test regex pattern
print("\nRegex Pattern Validation:")
email = "user@example.com"
email_pattern = "^[a-zA-Z0-9.!#$%&'*+/=?^_\\{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
print("Email:", email)
print("regex.match result:", regex.match(email, email_pattern))

print("\nSimple test completed!") 