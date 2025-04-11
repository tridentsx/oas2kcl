import .output
import file
import json

data: output.AS3_API_Request = json.decode(file.read(file.read_env("JSON_INPUTFILE")))
