package main

import "fmt"

func main() {
	client, err := NewInfraiClient()
	if err != nil {
		panic(err)
	}
	w := Workflow{Flags: client, Errors: client, Notifier: stdoutNotifier{}}
	result, err := w.Confirm(Appointment{ID: "apt-1042", PatientID: "patient-77", Channel: "sms"})
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
