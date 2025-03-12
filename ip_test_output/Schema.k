import regex
import net

schema Schema:
    ipv4?: str
    # ipv4 validation for ipv4
    ipv6?: str
    # ipv6 validation for ipv6
    # Format 'ipv6' validation is not directly supported

    check:
        ipv4 == None or net.is_IPv4(ipv4), "ipv4 must be a valid IPv4 address"
        ipv6 == None or net.is_IP(ipv6), "ipv6 must be a valid IP address"