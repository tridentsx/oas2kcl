schema Schema:
    coordinate: list

    check:
        # Min items: 3
        # Max items: 3
        coordinate != None
        coordinate == None or len(coordinate) >= 3
        coordinate == None or len(coordinate) <= 3

