# Changelog

## v.0.2.0
### Refactoring
- Moved internal packages to root directory.
- Removed commented lines and unreachable return in tests.
- Added codecov.io.
- Also added this change log file.

## v0.1.0
### Concurrent writers
- Added write.Distribute for writing concurrently. (closed #28)
- Used bytes.Buffer in tests instead of tmp files.
- Randomised error messages in tests to make sure there are no overlapping.
- Moved bootstrapping logic to internal/handler.
- Removed support for go 1.7 because that version doesn't support Shutdown.

## v0.0.5
### Eliminate race conditions
- Used buffers instead of writer.File to eliminate race conditions during tests.
- Made logger an implicit argument for handler.New .
- Deferred writer creation from main to handlers.
- Gracefully shut down the server. (closed #30)
- Handled SIGINT. (closed #18)

## v0.0.4
- Added log level that is received from payload.
- Concurrently writes the logs. (closed #23)
- Used a buffer for Plain.Read method for reducing calculations.

## v0.0.3
- Read configurations from file. (closed #21)
- Replaced flag library with go-flags.
- Handler receive multiple writers.
- Added handler errors file.
- Added a custom TextFormatter.
- Used logrus' formatter instead of the built in one for compatibility reasons. (closed #19)
- Only write a new line if the log entry doesn't have one.
