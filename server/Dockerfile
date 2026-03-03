# Start by building the application.
FROM golang:1.26 as build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -o /go/bin/app

# Now copy it into our base image.
FROM gcr.io/distroless/base-debian13:nonroot-amd64
COPY --from=build /go/bin/app /
EXPOSE 8080
ENTRYPOINT ["/app"]