import .output
import file
import json

data: output.DynatraceMonitoring = json.decode(file.read(file.read_env("JSON_INPUTFILE")))
