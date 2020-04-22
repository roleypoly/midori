# First stage: build the executable.
FROM --platform=$BUILDPLATFORM golang:1.14-alpine AS builder

ARG GOPROXY
ARG BUILDPLATFORM
ARG TARGETARCH
ARG TARGETOS
ENV GOPROXY ${GOPROXY}
ENV GOOS ${TARGETOS}
ENV GOARCH ${TARGETARCH}

ARG GIT_COMMIT
ARG GIT_BRANCH
ARG BUILD_DATE

# Create the user and group files that will be used in the running container to
# run the process as an unprivileged user.
RUN mkdir /user && \
  echo 'nobody:x:65534:65534:nobody:/:' > /user/passwd && \
  echo 'nobody:x:65534:' > /user/group

# Install the Certificate-Authority certificates for the app to be able to make
# calls to HTTPS endpoints.
# Git is required for fetching the dependencies.
RUN apk add --no-cache ca-certificates git

# Set the working directory outside $GOPATH to enable the support for modules.
WORKDIR /src

# Fetch dependencies first; they are less susceptible to change on every build
# and will therefore be cached for speeding up the next build
COPY ./go.mod ./go.sum ./
RUN go mod download

# Import the code from the context.
COPY ./ ./

# Build the executable to `/app`. Mark the build as statically linked.
RUN CGO_ENABLED=0 go build \
  -installsuffix "static" \
  -ldflags "-X github.com/roleypoly/midori/internal/version.GitCommit=${GIT_COMMIT} -X github.com/roleypoly/midori/internal/version.GitBranch=${GIT_BRANCH} -X github.com/roleypoly/midori/internal/version.BuildDate=${BUILD_DATE}" \
  -o /app ./cmd/midori

# Final stage: the running container.
FROM scratch AS final

# Import the user and group files from the first stage.
COPY --from=builder /user/group /user/passwd /etc/

# Import the Certificate-Authority certificates for enabling HTTPS.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Import the compiled executable from the first stage.
COPY --from=builder /app /app

# Declare the port on which the webserver will be exposed.
# As we're going to run the executable as an unprivileged user, we can't bind
# to ports below 1024.
EXPOSE 7669
EXPOSE 17669


# Perform any further action as an unprivileged user.
USER nobody:nobody

# Run the compiled binary.
ENTRYPOINT ["/app"]