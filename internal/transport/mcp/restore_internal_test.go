package mcpserver

import (
	"reflect"
	"testing"
)

// Putting the file back comes alone: any other argument of save_profile given
// beside it — one added later among them — makes the call ask for more than
// the adult agreed to.
func TestRestoreComesAloneWhateverElseIsGiven(t *testing.T) {
	t.Parallel()

	alone := saveProfileIn{Restore: true}
	if !alone.restoreAlone() {
		t.Fatal("restoreAlone() of restore alone = false, want true")
	}
	arguments := reflect.TypeFor[saveProfileIn]()
	for i := range arguments.NumField() {
		field := arguments.Field(i)
		if field.Name == "Restore" {
			continue
		}
		t.Run(field.Name, func(t *testing.T) {
			t.Parallel()

			in := saveProfileIn{Restore: true}
			given := reflect.ValueOf(&in).Elem().Field(i)
			switch given.Kind() {
			case reflect.Pointer:
				given.Set(reflect.New(given.Type().Elem()))
			case reflect.Slice:
				given.Set(reflect.MakeSlice(given.Type(), 0, 0))
			case reflect.Bool:
				given.SetBool(true)
			default:
				t.Fatalf("save_profile's argument %s is a %s, which this test does not know how to give", field.Name, given.Kind())
			}
			if in.restoreAlone() {
				t.Errorf("restoreAlone() with %s given = true, want false", field.Name)
			}
		})
	}
}
