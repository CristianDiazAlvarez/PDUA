name = pdua

build:
	@go build -o bin/$(name) src/**.go

# Linux install command
install: build
	@sudo cp bin/$(name) /usr/local/bin/$(name)
	@sudo chmod +x /usr/local/bin/$(name)