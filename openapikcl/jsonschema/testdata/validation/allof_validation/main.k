import .output
import file
import json

_data = json.decode(file.read(file.read_env("JSON_INPUTFILE")))
data = output.Pet {
    **_data
}
