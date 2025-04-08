import .output
import file
import json

data: output.NumberValidation = json.decode(file.read(file.read_env("JSON_INPUTFILE")))

print("Running for testcase=====> " + file.read_env("JSON_INPUTFILE"))
