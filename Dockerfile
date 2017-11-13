FROM golang:alpine

ADD peekaboo /peekaboo

ENTRYPOINT ["/peekaboo"]
