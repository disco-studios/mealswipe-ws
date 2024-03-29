#####################################################################
# Builder, download deps and prepare for build
##################################################################### 
FROM golang:stretch AS builder

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory /build
WORKDIR /build

# Get git configued to pull private
# TODO: Remove my PAT from here lol
RUN git config \
  --global \
  url."https://cameronlund4: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@github.com".insteadOf \
  "https://github.com"

# Copy and download dependency using go mod
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

#####################################################################
# Build, compile our image
# Separate from builder to save time, don't need to redownload deps
##################################################################### 
FROM builder AS build

# Move to working directory /build
WORKDIR /build

# Move the rest over
COPY . .

# # Build the application
RUN go build -o main ./cmd/websocket-server

# # Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# # Copy binary from build to main folder
RUN cp /build/main .

#####################################################################
# Prod, just our image and env vars
# Minimizes our prod image size
##################################################################### 
FROM --platform=linux/amd64 alpine:latest as prod

# Websocket server
EXPOSE 8080
# kubectl hooks
EXPOSE 8081

# elastic apm
ENV ELASTIC_APM_SERVICE_NAME=mealswipe-ws \
    ELASTIC_APM_SERVER_URL=https://9bfda03478fb43cf8552a3178e82e463.apm.us-east-1.aws.cloud.es.io:443 \
    ELASTIC_APM_SECRET_TOKEN=2nmMiRiidEavU8VXxJ

WORKDIR /root/
COPY --from=build /dist/main ./
ENTRYPOINT ["./main"]