+ 8 bytes -> read line by line

+ refactor:
    + inline
    + function
        + normal
        + return channel
    + class/struct
    + interface
    + embed/extension

+ http:
    ``````
    listener, err := net.Listen("tcp", ":8080")
    for {
        conn, err := accept()
        line := readLine(conn)
    }
    ``````
