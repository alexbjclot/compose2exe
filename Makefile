NAME = compose2exe
OUTPUT = dist
TARGETS = linux/amd64 linux/arm64 windows/amd64 darwin/amd64 darwin/arm64

os = $(word 1, $(subst /, ,$@))
arch = $(word 2, $(subst /, ,$@))

.PHONY: all build clean release

all: build

build:
	go build -o $(NAME) .

$(OUTPUT):
	mkdir -p $(OUTPUT)

release: $(OUTPUT) $(TARGETS)

$(TARGETS):
	GOOS=$(os) GOARCH=$(arch) go build -o "$(OUTPUT)/$(NAME)-$(os)-$(arch)" .

clean:
	rm -rf $(OUTPUT) $(NAME)
