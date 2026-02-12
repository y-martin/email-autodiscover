package template

import (
	"bytes"
	"text/template"
)

const thunderbirdTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<clientConfig version="1.1">
	<emailProvider id="{{.Domain}}">
	    <domain>{{.Domain}}</domain>

	    <displayName>%EMAILADDRESS%</displayName>
	    <displayShortName>%EMAILLOCALPART%</displayShortName>

	    <incomingServer type="imap">
			<hostname>{{.ImapHost}}</hostname>
			<port>993</port>
			<socketType>SSL</socketType>
			<authentication>password-cleartext</authentication>
			<username>%EMAILLOCALPART%</username>
		</incomingServer>

	    <outgoingServer type="smtp">
			<hostname>{{.SmtpHost}}</hostname>
			<port>587</port>
			<socketType>STARTTLS</socketType>
			<authentication>password-cleartext</authentication>
			<username>%EMAILLOCALPART%</username>
	    </outgoingServer>
	</emailProvider>
</clientConfig>`

func Thunderbird(args *Args) (string, error) {
	reportTmpl, err := template.New("report").Parse(thunderbirdTemplate)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	err = reportTmpl.Execute(&b, args)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
