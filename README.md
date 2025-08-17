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

+ http message:
    ``````
    start-line CRLF
    *( field-line CRLF )
    CRLF
    [ message-body ]

    // eg
    POST /users/primeagen HTTP/1.1
    Host: google.com
    CRLF
    {"name": "TheHTTPagen"}
    ``````

+ HTTP:
    + RFC 9110 – Covers HTTP "semantics."
    + RFC 9112 – Easier to read than RFC 7231, relies on understanding from RFC 9110.
