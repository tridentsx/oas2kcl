schema Schema:
    coordinate: list
    # Min items: 3
    # Max items: 3


    check:
        coordinate != None
        coordinate == None or len(coordinate) >= 3
        coordinate == None or len(coordinate) <= 3