package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
        "strings"

	"github.com/regnull/email-autodiscover/template"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

const (
       HTTP_PORT  = 80
       LISTEN_ADDRESS = "127.0.0.1"
)

type CmdArgs struct {
	HttpPort    int
        Address     string
	ConfigFile  string
        VerboseLogs bool
}

type Server struct {
	args *template.Args
}

func NewServer(templateArgs *template.Args) *Server {
	return &Server{args: templateArgs}
}

func (s *Server) HandleThunderbirdConfig(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("thunderbird config request")
	reply, err := template.Thunderbird(s.args)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate Thunderbird response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write([]byte(reply))
}

func (s *Server) HandleOutlookConfig(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("outlook config request")
	// Request comes as:
	// `<Autodiscover xmlns="https://schemas.microsoft.com/exchange/autodiscover/outlook/requestschema/2006">
	// <Request>
	//   <EMailAddress>user@contoso.com</EMailAddress>
	//   <AcceptableResponseSchema>https://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a</AcceptableResponseSchema>
	// </Request>
	// </Autodiscover>`

	if r.Method != "POST" {
		log.Warn().Msg("not POST request to outlook config")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Warn().Err(err).Msg("failed to read request body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var xmlRec struct {
		XMLName xml.Name `xml:"Autodiscover"`
		Request struct {
			XMLName                  xml.Name `xml:"Request"`
			EMailAddress             string   `xml:"EMailAddress"`
			AcceptableResponseSchema string   `xml:"AcceptableResponseSchema"`
		}
	}
	if err := xml.Unmarshal(body, &xmlRec); err != nil {
		log.Warn().Err(err).Msg("failed to parse request xml")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Debug().Str("schema", xmlRec.Request.AcceptableResponseSchema).Msg("got schema")
	newArgs := *s.args
        newArgs.EmailLocalPart = strings.Split(xmlRec.Request.EMailAddress, "@")[0]
	reply, err := template.OutlookMail(&newArgs)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate outlook response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-type", "text/xml")
	w.Write([]byte(reply))
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05", NoColor: true})

	var args CmdArgs
	flag.IntVar(&args.HttpPort, "http-port", 80, "HTTP port")
        flag.StringVar(&args.Address, "address", LISTEN_ADDRESS, "IP to bind to")
	flag.StringVar(&args.ConfigFile, "config", "", "config file")
        flag.BoolVar(&args.VerboseLogs, "verbose", false, "enable verbose logs")
        
	flag.Parse()

        if args.VerboseLogs {
	        zerolog.SetGlobalLevel(zerolog.DebugLevel)
        } else {
                zerolog.SetGlobalLevel(zerolog.InfoLevel)
        }

	if args.ConfigFile == "" {
		log.Fatal().Msg("--config-file must be specified")
	}

	log.Debug().Str("config-file", args.ConfigFile).Msg("using config file")
	config, err := ioutil.ReadFile(args.ConfigFile)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config file")
	}

	templateArgs := &template.Args{}
	err = yaml.Unmarshal(config, templateArgs)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse config file")
	}

	log.Info().Interface("config", templateArgs).Msg("running with")

	server := NewServer(templateArgs)
	http.HandleFunc("/mail/config-v1.1.xml", server.HandleThunderbirdConfig)
	http.HandleFunc("/Autodiscover/Autodiscover.xml", server.HandleOutlookConfig)

        log.Info().Str("address", args.Address).Int("port", args.HttpPort).Msg("starting http server...")
        err = http.ListenAndServe(fmt.Sprintf("%s:%d", args.Address, args.HttpPort), nil)
        if err != nil {
	    log.Fatal().Err(err).Msg("exiting")
	}
}
