
import .output
import file
import json

data: output.URITemplateValidation = json.decode(file.read(file.read_env("JSON_INPUTFILE")))
