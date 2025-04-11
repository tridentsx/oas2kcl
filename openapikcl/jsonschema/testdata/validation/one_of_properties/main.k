import output
import file
import json

data = json.decode(file.read(file.read_env("JSON_INPUTFILE")))
validation = output.Book {
    **data
}
