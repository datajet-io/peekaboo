FROM alpine:3.6

ADD peekaboo /peekaboo

ENTRYPOINT ["/peekaboo"]
