package mcpserver

import (
	"reflect"
	"testing"
)

// A change from the form is never taken for the choice of a topic on the card
// of a task: any field of the form sent beside the topic makes the change the
// form's — whichever field a later release adds to it.
func TestOnlyTheTopicAloneIsTheChoiceOnTheCard(t *testing.T) {
	t.Parallel()

	topic := "time.clocks"
	if alone := (editProfileIn{LessonTopic: &topic}); !alone.topicAlone() {
		t.Fatal("topicAlone() of the topic alone = false, want true")
	}
	fields := reflect.TypeFor[editProfileIn]()
	for i := range fields.NumField() {
		field := fields.Field(i)
		if field.Name == "LessonTopic" {
			continue
		}
		in := editProfileIn{LessonTopic: &topic}
		value := reflect.ValueOf(&in).Elem().Field(i)
		switch value.Kind() {
		case reflect.Pointer:
			value.Set(reflect.New(field.Type.Elem()))
		case reflect.Slice:
			value.Set(reflect.MakeSlice(field.Type, 0, 0))
		default:
			t.Fatalf("%s is a %s, which this check does not know how to send", field.Name, value.Kind())
		}
		if in.topicAlone() {
			t.Errorf("topicAlone() with %s sent beside the topic = true, want the change the form's", field.Name)
		}
	}
}
