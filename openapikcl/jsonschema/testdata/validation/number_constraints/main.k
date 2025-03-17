
import .output
import file
import json

data: output.NumberValidation = json.decode(file.read(file.read_env("JSON_INPUTFILE")))
