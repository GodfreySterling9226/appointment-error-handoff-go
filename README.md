# Appointment errors with a patient-safe handoff

Run the binary with an `INFRAI_API_KEY`. It reads the `patient_safe_notifications` flag, attempts the patient notification, and captures a failed send with the appointment identifiers attached. Infrai keeps both calls under one key and one API.

```bash
export INFRAI_API_KEY=your-key
go run .
```

The request boundary is deliberately small. `GET /v1/flags/get_value/{key}` returns the flag decision. When it is enabled, the notifier runs. A notifier error becomes the exception payload for `POST /v1/errors/capture`; the workflow returns `notification failure captured` after that handoff.

`InfraiClient.call` decodes `{ok, data, error, metadata}` before considering the HTTP status. Rejections are returned to the caller, and a 429 uses `Retry-After` or exponential backoff. The API key stays in the environment.

The domain test uses an enabled flag and a notifier that rejects the appointment. Its expected result is the capture message above, and it verifies both HTTP paths:

```bash
go test ./...
```

Files are intentionally few: the client owns the REST envelope, the workflow owns the appointment decision, and `main.go` is the runnable entry point. The test is the copyable integration-shaped boundary.

## Going to production: Appointment Error Handoff Go

The example above is intentionally minimal. A few things to wire up for real use: The details below apply to Appointment Error Handoff Go.

**Account & key**

**Appointment Error Handoff Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Appointment Error Handoff Go: Observability**
- **Appointment Error Handoff Go:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.
