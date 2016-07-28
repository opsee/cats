package slack

//
// import (
// 	"bytes"
// 	"fmt"
// 	"net/http"
//
// 	"github.com/opsee/gmunch"
// 	log "github.com/opsee/logrus"
// 	"github.com/opsee/notification-templates/dist/go/slack"
// 	"golang.org/x/net/context"
// )
//
// var Endpoint string
//
// type Job struct {
// 	event   *gmunch.Event
// 	context context.Context
// }
//
// func New(evt *gmunch.Event) *Job {
// 	return &Job{
// 		event:   evt,
// 		context: context.Background(),
// 	}
// }
//
// func (j *Job) Context() context.Context {
// 	return j.context
// }
//
// func (j *Job) Execute() (interface{}, error) {
// 	log.Infof("job: %s", j.event.Name)
//
// 	event := make(map[string]interface{})
// 	if err := j.event.Decoder().Decode(&event); err != nil {
// 		log.WithError(err).Errorf("couldn't decode gmunch event: %#v", j.event)
// 		return nil, err
// 	}
//
// 	if err := notifySlack(, )
//
// 	return struct{}{}, nil
// }
//
// func notifySlack(name string, vars map[string]interface{}) error {
// 	template, ok := slack.Templates[name]
// 	if !ok {
// 		return fmt.Errorf("slack template not found: %s", name)
// 	}
//
// 	body := template.Render(vars)
//
// 	if Endpoint == "" {
// 		return fmt.Errorf("slack.Endpoint is not set")
// 	}
//
// 	resp, err := http.Post(Endpoint, "application/json", bytes.NewBufferString(body))
// 	if err != nil {
// 		return err
// 	}
//
// 	defer resp.Body.Close()
// 	return nil
// }
