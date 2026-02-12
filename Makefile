all: email-autodiscover

email-autodiscover: cmd/autodiscover/main.go
	go build -o $@ $< 

clean:
	rm -f email-autodiscover