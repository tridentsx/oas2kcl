import Schema as schema

# Test with invalid IPv6 address - this should fail
invalid_instance = schema.Schema {
    ipv4 = "192.168.1.1"  # Valid IPv4
    ipv6 = "not-a-valid-ipv6"  # Invalid IPv6
}

print("If you see this, validation failed to catch the invalid IPv6 address") 