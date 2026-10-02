package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type failingNotifier struct{}

func (failingNotifier) Send(Appointment) error { return errors.New("sms provider rejected message") }

func TestConfirmCapturesNotificationFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/flags/get_value/patient_safe_notifications" {
			w.Write([]byte(`{"ok":true,"data":{"value":true}}`))
			return
		}
		if r.URL.Path == "/v1/errors/capture" && r.Method == "POST" {
			w.Write([]byte(`{"ok":true,"data":{}}`))
			return
		}
		t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	w := Workflow{Flags: &InfraiClient{BaseURL: server.URL, APIKey: "test", HTTP: server.Client()}, Errors: &InfraiClient{BaseURL: server.URL, APIKey: "test", HTTP: server.Client()}, Notifier: failingNotifier{}}
	got, err := w.Confirm(Appointment{ID: "apt-1", PatientID: "p-1", Channel: "sms"})
	if err != nil || got != "notification failure captured" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestConfirmSkipsNotificationWhenFlagIsMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"ok":false,"error":{"code":"FLAG_NOT_FOUND"}}`))
	}))
	defer server.Close()
	w := Workflow{Flags: &InfraiClient{BaseURL: server.URL, APIKey: "test", HTTP: server.Client()}, Errors: &InfraiClient{BaseURL: server.URL, APIKey: "test", HTTP: server.Client()}, Notifier: failingNotifier{}}
	got, err := w.Confirm(Appointment{ID: "apt-1", PatientID: "p-1", Channel: "sms"})
	if err != nil || got != "notification skipped" {
		t.Fatalf("got %q, %v", got, err)
	}
}
