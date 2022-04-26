package docformat

type MalformedDocumentError struct {
	err string
}

func (e *MalformedDocumentError) Error() string {
	return "malformed input: " + e.err
}

type InvalidArgumentError struct {
	Msg string
	Err error
}

func (e InvalidArgumentError) Error() string {
	return e.Msg
}

func (e InvalidArgumentError) Unwrap() error {
	return e.Err
}

func (e InvalidArgumentError) Is(target error) bool {
	_, ok := target.(InvalidArgumentError)
	return ok
}

type RequiredArgumentError struct {
	Msg string
	Err error
}

func (e RequiredArgumentError) Error() string {
	return e.Msg
}

func (e RequiredArgumentError) Unwrap() error {
	return e.Err
}

func (e RequiredArgumentError) Is(target error) bool {
	_, ok := target.(RequiredArgumentError)
	return ok
}

var (
	ErrEmptyBlock        = &MalformedDocumentError{"empty block"}
	ErrEmptyDoc          = &MalformedDocumentError{"empty document"}
	ErrEmptyConcept      = &MalformedDocumentError{"empty concept"}
	ErrEmptyNewsItem     = &MalformedDocumentError{"empty newsitem"}
	ErrEmptyAssignment   = &MalformedDocumentError{"empty assignment"}
	ErrEmptyList         = &MalformedDocumentError{"empty list"}
	ErrEmptyPlanningItem = &MalformedDocumentError{"empty planningitem"}
	ErrEmptyPackage      = &MalformedDocumentError{"empty list package"}
	ErrUnsupportedType   = &MalformedDocumentError{"unsuported type"}
)
