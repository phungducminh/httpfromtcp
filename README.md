+ 8 bytes -> read line by line

+ refactor:
    + inline
    + function
        + normal
        + return channel
        + when no states is required (pure functional)
    + class/struct
        + when containing states
    + interface
    + embed/extension

+ how to refactor:
    + avoid changing signatureof public method/interface/struct

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

+ Chapter 04: Request Lines
    + L3: Parsing a stream: we receive data in chunks and should be able to 
        parse it as it come in
    + how to wait until enough data in order to parse
        + if no available data, return (0, nil) denotes that not enough data
        + how to check if available data?
            + for request line: if no \r\n separater, return (0, nil)
        + do not return error as signal to request more data, return (0, nil) instead
        + why?
            + "Read(p []byte) (n)" func is supposed to return error if there is 
                actual error, caller should process n bytes before considering err
            + in order to allow this pattern work seamlessly, we will return 
                (0, nil) as nothing happened

        ``````
        req := newRequest()
        buf := make([]byte, 1024) 
        bufIdx := 0

        // buf[:bufIdx] denote data has been read and available to parse into request
        for !req.done() {
            n, err := rd.Read(buf[bufIdx:])
            bufIdx += n
            readN, err := req.parse(buf[:bufIdx])
            // shift buffer to left
        }

        func (r *Request) parse(p []byte) (int, error) {
            if not_enough_data(p) {
                return 0, nil
            }

            return parse(p), nil
        }
        ``````

``````
 HTTP-message   = start-line CRLF
                  *( field-line CRLF )
                  CRLF
                  [ message-body ]
``````

+ Test body:
    ``````
    curl -X POST http://localhost:42069/coffee \
    -H 'Content-Type: application/json' \
    -d '{"type": "dark mode", "size": "medium"}'
    ``````

+ why changing from 
    ``````
    type Handler func(w io.Writer, req *request.Request) *HandlerError"
    ``````
    to
    ``````
    type Handler func(w *response.Writer, req *request.Request)
    ``````
    will result into more flexible function signature for custom handlers?
    + allow client to modify headers, status code, ..., client will decide what 
        the response will be 


+ how to use netcat to send http request?
    ``````
    echo -e "GET /httpbin/stream/5 HTTP/1.1\r\nHost: localhost:42069\r\nConnection: close\r\n\r\n" | nc localhost 42069
    ``````
