import .output
import file
import json

data: output.Schema = json.decode(file.read(file.read_env("JSON_INPUTFILE")))
