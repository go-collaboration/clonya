.POSIX:

PREFIX?=/usr/local
BINDIR?=$(PREFIX)/bin
GO?=go
BUILD_OPTS?=-trimpath

SOURCES = $(shell find . -name '*.go')
MODULE_SOURCES = go.mod go.sum

all: clonya

clonya: $(MODULE_SOURCES) $(SOURCES)
	$(GO) build $(BUILD_OPTS) -o clonya

clean:
	rm -rf clonya

install: clonya
	install -Dm755 clonya $(DESTDIR)$(BINDIR)/clonya

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/clonya

.PHONY: all clean install uninstall
