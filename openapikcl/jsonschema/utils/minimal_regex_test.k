import regex

# Test with a very simple pattern
pattern1 = r"^test$"
print("Pattern1:", pattern1)
print("Match 'test':", regex.match("test", pattern1))
print("Match 'test123':", regex.match("test123", pattern1))

# Test with simple date pattern
pattern2 = r"^2024-03-20$"
print("Pattern2:", pattern2)
print("Match '2024-03-20':", regex.match("2024-03-20", pattern2))

# Test with escaped digit pattern for date
pattern3 = r"^\\d{4}-\\d{2}-\\d{2}$"
print("Pattern3:", pattern3)
print("Match '2024-03-20':", regex.match("2024-03-20", pattern3))

# Test with unescaped digit pattern for date
pattern4 = r"^\d{4}-\d{2}-\d{2}$"
print("Pattern4:", pattern4)
print("Match '2024-03-20':", regex.match("2024-03-20", pattern4)) 