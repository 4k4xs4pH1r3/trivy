FROM alpine:3.13.7
RUN apk --no-cache add ca-certificates git rpm
COPY trivy /usr/local/bin/trivy
COPY contrib/*.tpl contrib/
ENTRYPOINT ["trivy"]
