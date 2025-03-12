import Schema as schema

# Test with invalid IP addresses - this should fail
invalid_instance = schema.Schema {
    ipv4 = "192.168.1.300"  # Invalid IPv4
}

print("If you see this, validation failed to catch the invalid IP address") 