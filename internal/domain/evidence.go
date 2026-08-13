package domain

import "time"

// EvidenceReference points at attributable source material without treating the
// source as canonical Calathea authority.
type EvidenceReference struct {
	id        EvidenceReferenceID
	locator   string
	createdAt time.Time
}

func NewEvidenceReference(id EvidenceReferenceID, locator string, createdAt time.Time) (EvidenceReference, error) {
	if err := requireIdentifier("evidence reference id", string(id)); err != nil {
		return EvidenceReference{}, err
	}
	if err := requireText("evidence locator", locator); err != nil {
		return EvidenceReference{}, err
	}
	if createdAt.IsZero() {
		return EvidenceReference{}, errZeroTime("evidence reference created at")
	}
	return EvidenceReference{id: id, locator: locator, createdAt: createdAt}, nil
}

func (e EvidenceReference) ID() EvidenceReferenceID { return e.id }
func (e EvidenceReference) Locator() string         { return e.locator }
func (e EvidenceReference) CreatedAt() time.Time    { return e.createdAt }

// TraceEntry is one structured explanation step. Codes are stable machine-facing
// semantics; messages are human-facing detail.
type TraceEntry struct {
	code        string
	message     string
	evidenceIDs []EvidenceReferenceID
}

func NewTraceEntry(code, message string, evidenceIDs []EvidenceReferenceID) (TraceEntry, error) {
	if err := requireText("trace code", code); err != nil {
		return TraceEntry{}, err
	}
	if err := requireText("trace message", message); err != nil {
		return TraceEntry{}, err
	}
	for _, id := range evidenceIDs {
		if err := requireIdentifier("trace evidence reference id", string(id)); err != nil {
			return TraceEntry{}, err
		}
	}
	return TraceEntry{code: code, message: message, evidenceIDs: cloneEvidenceIDs(evidenceIDs)}, nil
}

func (e TraceEntry) Code() string                       { return e.code }
func (e TraceEntry) Message() string                    { return e.message }
func (e TraceEntry) EvidenceIDs() []EvidenceReferenceID { return cloneEvidenceIDs(e.evidenceIDs) }

// OperationTrace is the shared immutable explanation container for one operation.
type OperationTrace struct {
	operationID OperationID
	entries     []TraceEntry
}

func NewOperationTrace(operationID OperationID, entries []TraceEntry) (OperationTrace, error) {
	if err := requireIdentifier("operation id", string(operationID)); err != nil {
		return OperationTrace{}, err
	}
	return OperationTrace{operationID: operationID, entries: cloneTraceEntries(entries)}, nil
}

func (t OperationTrace) OperationID() OperationID { return t.operationID }
func (t OperationTrace) Entries() []TraceEntry    { return cloneTraceEntries(t.entries) }

func cloneEvidenceIDs(values []EvidenceReferenceID) []EvidenceReferenceID {
	if len(values) == 0 {
		return nil
	}
	return append([]EvidenceReferenceID(nil), values...)
}

func cloneTraceEntries(values []TraceEntry) []TraceEntry {
	if len(values) == 0 {
		return nil
	}
	result := make([]TraceEntry, len(values))
	for i, value := range values {
		result[i] = TraceEntry{code: value.code, message: value.message, evidenceIDs: cloneEvidenceIDs(value.evidenceIDs)}
	}
	return result
}
