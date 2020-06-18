coverage.txt:
	go test -coverprofile=coverage.txt -covermode=atomic -timeout 5s .

test: coverage.txt
	golint -set_exit_status .
	go test -race .
	golangci-lint run -E asciicheck -E bodyclose -E depguard -E dogsled -E dupl -E gochecknoinits -E goconst -E gocritic -E godot -E godox -E gofmt -E goimports -E golint -E gomodguard -E goprintffuncname -E interfacer -E maligned -E misspell -E nolintlint -E prealloc -E rowserrcheck -E stylecheck -E unconvert -E unparam -E whitespace -E wsl

bench:
	go test -bench=.

html: coverage.txt
	go tool cover -html=coverage.txt
