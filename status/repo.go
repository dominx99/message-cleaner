package status_repo

import "github.com/nlopes/slack"

type Status struct {
	Name    string
	User    string
	Timeout int64 `default0:"3600"`
}

func (s *Status) Set(api *slack.Client) error {
	var err error

	switch s.Name {
	case "working":
		err = api.SetUserCustomStatusWithUser(s.User, "Working", ":workingonit:", s.Timeout)
	case "end":
		err = api.SetUserCustomStatusWithUser(s.User, "Ended work", ":disappear:", s.Timeout)
	case "break":
		err = api.SetUserCustomStatusWithUser(s.User, "Break", ":outofoffice:", s.Timeout)
	case "eat":
		err = api.SetUserCustomStatusWithUser(s.User, "Eating break", ":chompy:", s.Timeout)
	default:
		err = api.SetUserCustomStatusWithUser(s.User, s.Name, ":"+s.Name+":", s.Timeout)
	}

	return err
}
