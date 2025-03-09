import regex

schema UserProfile:
    # A user profile with various string constraints
    # User roles
    roles: list[str]
    # Unique items: true
    # Item allowed values: "admin", "moderator", "user", "guest"

    # Email address
    email: str
    # Format: email

    # User tags (alphanumeric and hyphens only)
    tags: list[str]
    # Item min length: 1
    # Item max length: 20
    # Item pattern: ^[a-zA-Z0-9-]+$

    # Additional email addresses
    emails: list[str]
    # Unique items: true
    # Item format: email

    socialProfiles: UserProfileSocialProfiles
    addresses: list[UserProfileAddress]
    # Username (alphanumeric and underscores only)
    username: str
    # Min length: 3
    # Max length: 20
    # Regex pattern: ^[a-zA-Z0-9_]+$

    # User's biography
    bio: str
    # Max length: 500

    # User's birthdate (YYYY-MM-DD)
    birthdate: str
    # Format: date

    # User's last login timestamp
    lastLogin: str
    # Format: date-time

    # Phone number in E.164 format
    phone: str
    # Regex pattern: ^\+?[1-9]\d{1,14}$

    # User's display name
    displayName: str
    # Min length: 1
    # Max length: 50

    # Password (must contain at least one uppercase letter, one lowercase letter, one number, and one special character)
    password: str
    # Min length: 8
    # Max length: 100
    # Regex pattern: ^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$

    # User's website URL
    website: str
    # Format: uri

    # User's unique identifier (UUID)
    userId: str
    # Format: uuid

    # Account status
    status: str
    # Allowed values: "active", "inactive", "suspended", "pending"


schema UserProfileSocialProfiles:
    # Twitter handle
    twitter: str
    # Facebook profile URL
    facebook: str
    # LinkedIn profile URL
    linkedin: str
    # GitHub username
    github: str

schema UserProfileAddress:
    country: str
    # Min length: 2
    # Max length: 2
    # Regex pattern: ^[A-Z]{2}$
    line1: str
    # Min length: 1
    # Max length: 100
    line2: str
    # Max length: 100
    city: str
    # Min length: 1
    # Max length: 100
    postalCode: str
    # Regex pattern: ^[A-Z0-9]{1,10}$


# Validator schema for UserProfileAddress
schema UserProfileAddressValidator:
    self: UserProfileAddress

    check:
        self.country == None or len(self.country) >= 2, "country length must be at least 2"
        self.country == None or len(self.country) <= 2, "country length must be at most 2"
        self.country == None or regex.match("^[A-Z]{2}$", self.country), "country must match pattern"
        self.line1 == None or len(self.line1) >= 1, "line1 length must be at least 1"
        self.line1 == None or len(self.line1) <= 100, "line1 length must be at most 100"
        self.line2 == None or len(self.line2) <= 100, "line2 length must be at most 100"
        self.city == None or len(self.city) >= 1, "city length must be at least 1"
        self.city == None or len(self.city) <= 100, "city length must be at most 100"
        self.postalCode == None or regex.match("^[A-Z0-9]{1,10}$", self.postalCode), "postalCode must match pattern"


# Validator schema for UserProfile
schema UserProfileValidator:
    self: UserProfile

    check:
        self.status == None or self.status in ["active", "inactive", "suspended", "pending"], "status must be one of the allowed values"
        self.username == None or len(self.username) >= 3, "username length must be at least 3"
        self.username == None or len(self.username) <= 20, "username length must be at most 20"
        self.username == None or regex.match("^[a-zA-Z0-9_]+$", self.username), "username must match pattern"
        self.bio == None or len(self.bio) <= 500, "bio length must be at most 500"
        self.phone == None or regex.match("^\+?[1-9]\d{1,14}$", self.phone), "phone must match pattern"
        self.displayName == None or len(self.displayName) >= 1, "displayName length must be at least 1"
        self.displayName == None or len(self.displayName) <= 50, "displayName length must be at most 50"
        self.password == None or len(self.password) >= 8, "password length must be at least 8"
        self.password == None or len(self.password) <= 100, "password length must be at most 100"
        self.password == None or regex.match("^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$", self.password), "password must match pattern"