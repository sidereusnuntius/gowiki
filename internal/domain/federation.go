package domain

import "net/url"

// FedObj is an internal type used to represent arbitrary Activitystreams objects.
type FedObj struct {
	// Iri is the object's unique identifier.
	Iri *url.URL
	// RawJSON is the object's JSON-LD representation, which may be empty, for instance, when storing a collection
	// in the database.
	RawJSON    string
	ApType     string
	Local      bool
	LocalTable string
	LocalId    int64
}
