package controllers

import (
	"encoding/json"
	"github.com/robfig/revel"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var audience string

type Persona struct {
	*revel.Controller
	UserEmail *string
}

type personaResponse struct {
	Status string `json:"status"`
	Email *string `json:"email,omitempty"`
	Audience *string `json:"audience,omitempty"`
	Expires *int64 `json:"expires,omitempty"`
	Issuer *string `json:"issuer,omitempty"`
	Reason *string `json:"reason,omitempty"`
}

type personaRA struct {
	Email string
}

type ErrorString string
func (s ErrorString) Error() string {
	return string(s)
}

func (c Persona) Login(assertion string, redirect string) revel.Result {
	resp, err := http.PostForm("https://verifier.login.persona.org/verify", url.Values{
		"assertion": {assertion},
		"audience": {audience},
	})
	if err != nil {
		revel.WARN.Fatal("Failed to verify assertion: %s", err)
		return c.RenderError(err)
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	pr := personaResponse{}
	if err := decoder.Decode(&pr); err != nil {
		revel.WARN.Fatal("Failed to decode JSON")
	}
	if pr.Status == "okay" {
		c.Session["persona/email"] = *pr.Email
		c.Session["persona/exp"] = strconv.FormatInt(*pr.Expires, 36)
		if redirect == "" {
			return c.Render(pr)
		}
		return c.Redirect(redirect)
	}
	return c.RenderError(ErrorString(*pr.Reason))
}

func (c Persona) Logout(redirect string) revel.Result {
	delete(c.Session, "persona/email")
	delete(c.Session, "persona/exp")
	if redirect != "" {
		return c.Redirect(redirect)
	}
	return c.RenderText("Logged out")
}

func (p Persona) CheckUser() revel.Result {
	var ok bool
	var exp, email string
	p.UserEmail = nil
	if exp, ok = p.Session["persona/exp"]; !ok {
		return nil
	}
	if email, ok = p.Session["persona/email"]; !ok {
		return nil
	}

	revel.ERROR.Print("foo")
	if expms, err := strconv.ParseInt(exp, 36, 64); err != nil {
		revel.ERROR.Fatal("Failed to parse expiration: %s", err)
	} else {
		expt := time.Unix(expms / 1000, (expms % 1000) * 1000000)
		if expt.Before(time.Now()) {
			p.Logout("")
			return nil
		}
	}
	p.UserEmail = &email
	p.RenderArgs["persona"] = personaRA{
		Email: email,
	}
	revel.WARN.Print(p.RenderArgs)
	return nil
}

func init() {
	revel.OnAppStart(func() {
		var found bool
		if audience, found = revel.Config.String("persona.audience"); !found {
			revel.ERROR.Fatal("No persona.audience found")
		}
	})

	revel.InterceptMethod((*Persona).CheckUser, revel.BEFORE)
}
