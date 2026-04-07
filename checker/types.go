// Package checker implements a dummy checker for happyDomain.
//
// This is an educational example that demonstrates all the building blocks
// needed to create a happyDomain checker. It performs no real monitoring;
// instead, it returns a configurable message and a random score, so you can
// focus on the structure without worrying about external dependencies.
package checker

import "time"

// ObservationKeyDummy is the unique key that identifies observations
// produced by this checker. Every checker must define at least one key so
// happyDomain can store and retrieve its data.
const ObservationKeyDummy = "dummy"

// DummyData is the data structure returned by Collect.
//
// When happyDomain collects an observation, it serialises this struct to JSON
// and stores it. Later, during evaluation, the same JSON is deserialised back
// into this struct. Design this type to hold everything your rules will need
// to decide OK / Warning / Critical.
type DummyData struct {
	// Message is an arbitrary string returned as part of the observation.
	Message string `json:"message"`

	// Score is a number between 0 and 100. The evaluation rules compare it
	// against user-defined thresholds to determine the check status.
	Score float64 `json:"score"`

	// CollectedAt records when the observation was taken. It is used by the
	// metrics reporter to timestamp the extracted metrics.
	CollectedAt time.Time `json:"collected_at"`
}
