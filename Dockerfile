FROM golang
ADD . /go/src/github.com/abacusresearch/android-release-bot
RUN go get github.com/abacusresearch/android-release-bot/...
RUN go install github.com/abacusresearch/android-release-bot
ENTRYPOINT /go/bin/android-release-bot
