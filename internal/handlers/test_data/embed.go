package test_data

import (
	_ "embed"
)

//go:embed creation-event.json
var CreationEventJSONFile []byte

//go:embed interruption-event.json
var InterruptionEventJSONFile []byte
