package services

import (
	"errors"
	"time"
	"tools/runtimes/config"
	suggestions "tools/runtimes/db/Suggestions"
	"tools/runtimes/funcs"
	"tools/runtimes/requests"

	"github.com/tidwall/gjson"
)

func AddSuggestion(sug *suggestions.Suggestion) error {
	r, err := requests.New(&requests.Config{Timeout: time.Second * 30})
	if err != nil {
		return err
	}

	hd := funcs.ServerHeader(config.VERSION, config.VERSIONCODE)
	body, err := config.Json.Marshal(sug)
	if err != nil {
		return err
	}

	resp, err := r.Post("add_sug", body, hd)
	if err != nil {
		return err
	}
	gs := gjson.ParseBytes(resp)
	if gs.Get("code").Int() != 200 {
		return errors.New(gs.Get("msg").String())
	}
	return nil
}
