import Schema as schema

# Test with valid IP addresses
valid_instance = schema.Schema {
    ipv4 = "192.168.1.1"
    ipv6 = "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
}

print("Valid instance:", valid_instance)

# Note: Testing invalid IP addresses would cause validation errors
# and stop execution, so we can't test them in the same file
