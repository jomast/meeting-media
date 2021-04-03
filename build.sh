#/bin/sh

# build for amd64 and arm64
CGO_ENABLED=1 GOARCH=amd64 go build -ldflags="-s -w" -o meeting-media-amd64
CGO_ENABLED=1 GOARCH=arm64 go build -ldflags="-s -w" -o meeting-media-arm64

# merge both bins together to make universal binary
lipo -create -output meeting-media meeting-media-arm64 meeting-media-amd64

# create MacOS app
appify -author "Jonathan Stanford" -icon icon.png -id "MeetingMedia" -name "MeetingMedia" -version "1.1" ./meeting-media
