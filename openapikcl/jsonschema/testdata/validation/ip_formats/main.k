
import .output
import file
import json

data: output.IPv4Validator = json.decode(file.read(file.read_env("JSON_INPUTFILE")))
