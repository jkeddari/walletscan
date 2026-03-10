# Build-Stage
FROM golang:1.26-alpine AS build
WORKDIR /app
COPY . .
RUN go generate ./... &&  go build -o api ./cmd/api


FROM gcr.io/distroless/base
COPY --from=build /app/assets /assets
COPY --from=build /app/api /api
COPY --from=build /app/coinlist.json /coinlist.json
CMD ["/api"]
