package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"
)

type config struct {
	StripePublicKey string `json:"stripe_public"`
	StripeSecretKey string `json:"stripe_secret"`
	Port            int
}

func (c config) Address() string {
	return fmt.Sprintf(":%d", c.Port)
}

func Config() (conf config) {
	b, err := ioutil.ReadFile("./settings.json")
	if err != nil {
		log.Panic(err)
	}
	if err = json.Unmarshal(b, &conf); err != nil {
		log.Panic(err)
	}
	if conf.StripePublicKey == "" {
		log.Panic("Stripe public key is missing from settings.json")
	}
	if conf.StripeSecretKey == "" {
		log.Panic("Stripe secret key is missing from settings.json")
	}
	if conf.Port == 0 {
		log.Panic("Port is missing from settings.json")
	}
	return
}

// Templates
var rootTemplate = template.Must(template.ParseFiles("./templates/root.html"))

// Handlers
func favicon(w http.ResponseWriter, r *http.Request) {
	return
}

type server struct {
	config config
}

func (srv server) root(w http.ResponseWriter, r *http.Request) {
	attrs := map[string]interface{}{
		"Public": srv.config.StripePublicKey,
	}
	if err := rootTemplate.Execute(w, attrs); err != nil {
		log.Panic(err)
	}
}

// Charge endpoint
func (srv server) charge(w http.ResponseWriter, r *http.Request) {
	// Other tokens: stripeTokenType, stripeEmail
	token := r.PostFormValue("stripeToken")

	params := stripe.ChargeParams{
		Amount:   1000, // in cents
		Desc:     "Test charge",
		Currency: "USD",
	}
	if err := params.SetSource(token); err != nil {
		fmt.Fprintf(w, "SetSource failed: %s", err)
		return
	}

	if _, err := charge.New(&params); err != nil {
		fmt.Fprintf(w, "Charge failed: %s", err)
		return
	}
	fmt.Fprintf(w, "Charge succeeded!")
}

func Server(conf config) server {
	return server{config: conf}
}

func main() {
	// Read the local settings
	conf := Config()

	// Set the Stripe secret token server-side
	stripe.Key = conf.StripeSecretKey

	// Attach the handlers
	srv := Server(conf)
	http.HandleFunc("/favicon.ico", favicon)
	http.HandleFunc("/", srv.root)
	http.HandleFunc("/charge", srv.charge)

	// Start the server
	log.Printf("Starting server on %s", conf.Address())
	http.ListenAndServe(conf.Address(), nil)
}
