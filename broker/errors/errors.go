package errors

import "fmt"

// Error is the type that implements the error interface.
type Error struct {
	// Kind is the class of error, such as permission failure, or "Other" if its
	// class is unknown or irrelevant.
	Kind Kind

	// The underlying error that triggered this one, if any.
	Err error
}

// Kind defines the kind of error this is.
type Kind uint8

// Kinds of errors.
const (
	//
	// General Error Codes
	//
	GENERR001 Kind = iota // The Message Body is not in the expected format, for example mandatory fields are missing.
	GENERR002             // The provided messageType is not supported.
	GENERR003             // The expiration date of the Message had passed at the point at which delivery was attempted.
	GENERR004             // Invalid, missing or corrupt headers were detected on the Message.
	GENERR005             // Maximum number of connection retries exceeded when attempting to send the Message.
	GENERR006             // An error occurred interacting with the underlying system.
	GENERR007             // Malformed JSON was detected in the Message Body.
	GENERR008             // An attempt to roll back a transaction failed.
	GENERR009             // An unexpected or unknown error occurred.
	GENERR010             // Maximum number of Message resend retries exceeded.

	//
	// Application Error Codes
	//

	// Metadata Error Codes
	APPERRMET001 // Received a Metadata UPDATE with a datasetUuid that does not exist.
	APPERRMET002 // Received a Metadata DELETE with a datasetUuid that does not exist.

	// Vocabulary Error Codes
	APPERRVOC001 // Received a Vocabulary UPDATE with a vocabularyId that does not exist.
	APPERRVOC002 // Received a Vocabulary DELETE with a vocabularyId that does not exist.

	// Term Error Codes
	APPERRTER001 // Received a Term UPDATE with a vocabularyId that does not exist.
	APPERRTER002 // Received a Term UPDATE with a termId that does not exist in the given vocabularyId.
	APPERRTER003 // Received a Term DELETE with a vocabularyId that does not exist.
	APPERRTER004 // Received a Term DELETE with a termId that does not exist in the given voabularyId.
)

// Error implements error
func (e *Error) Error() string {
	var ret string
	switch e.Kind {
	case GENERR001:
		ret = "GENERR001"
	case GENERR002:
		ret = "GENERR002"
	case GENERR003:
		ret = "GENERR003"
	case GENERR004:
		ret = "GENERR004"
	case GENERR005:
		ret = "GENERR005"
	case GENERR006:
		ret = "GENERR006"
	case GENERR007:
		ret = "GENERR007"
	case GENERR008:
		ret = "GENERR008"
	case GENERR009:
		ret = "GENERR009"
	case GENERR010:
		ret = "GENERR010"

	case APPERRMET001:
		ret = "APPERRMET001"
	case APPERRMET002:
		ret = "APPERRMET002"

	case APPERRVOC001:
		ret = "APPERRVOC001"
	case APPERRVOC002:
		ret = "APPERRVOC002"

	case APPERRTER001:
		ret = "APPERRTER001"
	case APPERRTER002:
		ret = "APPERRTER002"
	case APPERRTER003:
		ret = "APPERRTER003"
	case APPERRTER004:
		ret = "APPERRTER004"
	}

	return fmt.Sprintf("[%s] %s", ret, e.Err)
}
