############################
# STEP 1 build executable binary
############################
FROM golang:1.14.0-alpine3.11 as builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Create appuser
ENV USER=appuser
ENV UID=10001

# See https://stackoverflow.com/a/55757473/12429735
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"
WORKDIR $GOPATH/src/mypackage/myapp/

# Fetch dependencies.
COPY go.* ./
RUN go mod download

COPY . ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
      -ldflags='-w -s -extldflags "-static"' -a \
      -o /go/bin/dir2consul .

############################
# STEP 2 build a small image
############################
FROM scratch

ARG BUILD_DATE
ARG VCS_REF
ARG VCS_URL
ARG VERSION

LABEL "maintainer"="Empower Rangers <empower-rangers@code42.com>"
LABEL "org.label-schema.schema-version"="1.0"
LABEL "org.label-schema.build-date"=$BUILD_DATE
LABEL "org.label-schema.vcs-ref"=$VCS_REF
LABEL "org.label-schema.vcs-url"=$VCS_URL
LABEL "org.label-schema.version"=$VERSION

# Import from builder.
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy our static executable
COPY --from=builder /go/bin/dir2consul /go/bin/dir2consul

# Use an unprivileged user.
USER appuser:appuser

# Run the binary.
ENTRYPOINT ["/go/bin/dir2consul"]