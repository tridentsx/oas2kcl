# No schema imports needed - schemas in same directory

schema Schema:
    name?: str
    socket?: {str:int} = {HTTPS: 443, HTTP: 80}