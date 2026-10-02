package main

import "fmt"

type Appointment struct{ ID, PatientID, Channel string }

type Notifier interface{ Send(Appointment) error }

type Workflow struct {
	Flags    *InfraiClient
	Errors   *InfraiClient
	Notifier Notifier
}

func (w Workflow) Confirm(a Appointment) (string, error) {
	enabled, err := w.Flags.FlagValue("patient_safe_notifications")
	if err != nil {
		return "", err
	}
	if !enabled {
		return "notification skipped", nil
	}
	if err := w.Notifier.Send(a); err != nil {
		if captureErr := w.Errors.CaptureAppointmentError(a, err); captureErr != nil {
			return "", captureErr
		}
		return "notification failure captured", nil
	}
	return fmt.Sprintf("appointment %s confirmed and patient notified", a.ID), nil
}

type stdoutNotifier struct{}

func (stdoutNotifier) Send(a Appointment) error { return nil }
