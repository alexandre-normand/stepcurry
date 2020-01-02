# Cloud Functions deployment

## Requirements

[landler](https://github.com/alexandre-normand/landler) must be installed via `go install github.com/alexandre-normand/landler`

## Deploy

```
go mod init github.com/alexandre-normand/rogerchallenger/cloudfunctions
go build .
go mod vendor
landler | xargs -I % gcloud functions deploy % --runtime go111 --trigger-http --allow-unauthenticated
```
