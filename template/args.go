package template

type Args struct {
	Domain          string `yaml:"domain"`
	ImapHost        string `yaml:"imap-host"`
	SmtpHost        string `yaml:"smtp-host"`
        EmailLocalPart  string `json:"-"` // left part of email
}
