package status_repo

import "github.com/nlopes/slack"

type Status struct {
	Name    string
	Timeout int64 `default0:"3600"`
}

func (s *Status) Set(api *slack.Client) error {
	var err error

	switch s.Name {
	case "working":
		err = api.SetUserCustomStatus("Working", ":workingonit:", s.Timeout)
	case "end":
		err = api.SetUserCustomStatus("Ended work", ":disappear:", s.Timeout)
	case "break":
		err = api.SetUserCustomStatus("Break", ":outofoffice:", s.Timeout)
	case "eat":
		err = api.SetUserCustomStatus("Eating break", ":chompy:", s.Timeout)
	default:
		err = api.SetUserCustomStatus(s.Name, ":"+s.Name+":", s.Timeout)
	}

	return err
}
