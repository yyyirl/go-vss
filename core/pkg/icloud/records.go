package icloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

//go:generate ../bin/stringer -type=OperationType -linecomment -output=records_string.go

// OperationType is the type of an operation.
type OperationType uint8

const (
	// Create a new record. This operation fails if a record with the same
	// record name already exists.
	Create OperationType = iota + 1 // create
	// Update an existing record. Only the fields specified are changed.
	Update // update
	// ForceUpdate updates an existing record regardless of conflicts. Creates a
	// record if it doesn’t exist.
	ForceUpdate // forceUpdate
	// Replace a record with the specified record. The fields whose values are
	// not specified are set to null.
	Replace // replace
	// ForceReplace replaces a record with the specified record regardless of
	// conflicts. Creates a record if it doesn’t exist.
	ForceReplace // forceReplace
	// Delete the specified record.
	Delete // delete
	// ForceDelete deletes the specified record regardless of conflicts.
	ForceDelete // forceDelete
)

// MarshalJSON implements json.Marshaler. It is in place to marshal the
// OperationType to its string representation because that's what the server
// expects.
func (ot OperationType) MarshalJSON() ([]byte, error) {
	return json.Marshal(ot.String())
}

// UnmarshalJSON implements json.Unmarshaler. It is in place to unmarshal the
// OperationType from the string representation the server returns.
func (ot *OperationType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch s {
	case Create.String():
		*ot = Create
	case Update.String():
		*ot = Update
	case ForceUpdate.String():
		*ot = ForceUpdate
	case Replace.String():
		*ot = Replace
	case ForceReplace.String():
		*ot = ForceReplace
	case Delete.String():
		*ot = Delete
	case ForceDelete.String():
		*ot = ForceDelete
	default:
		return fmt.Errorf("unknown operation type %q", s)
	}

	return nil
}

// RecordsRequest is the request to every operation of the RecordsService.
type RecordsRequest struct {
	// Operations to apply to records in the database. Limited to
	// MaxOperationPerRequest.
	Operations []RecordOperation `json:"operations,omitempty"`
}

// RecordOperation is an operation on a single record.
type RecordOperation struct {
	// Type of the operation.
	Type OperationType `json:"operationType,omitempty"`
	// Record to create, update, replace or delete.
	Record Record `json:"record,omitempty"`
}

// Record is a record in the database.
type Record struct {
	// Name of the record.
	Name string `json:"recordName,omitempty"`
	// Type of the record.
	Type string `json:"recordType,omitempty"`
	// Fields of the record.
	Fields          Fields `json:"fields,omitempty"`
	RecordChangeTag string `json:"recordChangeTag,omitempty"`

	ServerErrorCode string `json:"serverErrorCode"`
	Reason          string `json:"reason"`
}

// Fields is a list of fields.
type Fields []Field

// MarshalJSON implements json.Marshaler. It is in place to marshal the
// Fields as a JSON object because that's what the server expects.
func (f Fields) MarshalJSON() ([]byte, error) {
	fields := make(map[string]*Field, len(f))

	for i, field := range f {
		fields[field.Name] = &f[i]
	}

	return json.Marshal(fields)
}

// UnmarshalJSON implements json.Unmarshaler. It is in place to unmarshal the
// fields JSON object returned by the server into a proper Fields value.
func (f *Fields) UnmarshalJSON(b []byte) error {
	fields := make(map[string]*Field)

	if err := json.Unmarshal(b, &fields); err != nil {
		return err
	}

	for fieldName, field := range fields {
		field.Name = fieldName
		*f = append(*f, *field)
	}

	return nil
}

// A Field is part of a record.
type Field struct {
	// Name of the field.
	Name string `json:"-"`
	// Type of the field.
	Type string `json:"type,omitempty"`
	// Value of the field.
	Value interface{} `json:"value,omitempty"`
}

// RecordsResponse is the response recevied from every operation of the
// RecordsService.
type RecordsResponse struct {
	Records []Record `json:"records,omitempty"`
}

// RecordsService handles communication with the record related operations of
// the CloudKit Web Services API.
//
// CloudKit Web Services Reference: https://developer.apple.com/library/archive/documentation/DataManagement/Conceptual/CloudKitWebServicesReference/ModifyRecords.html
type RecordsService service

// Modify records in a database.
func (s *RecordsService) Modify(ctx context.Context, database Database, req RecordsRequest) (*RecordsResponse, error) {
	path := "/" + database.String() + s.basePath

	var res RecordsResponse
	if err := s.client.call(ctx, http.MethodPost, path, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

type AssetUploadToken struct {
	RecordName string `json:"recordName,omitempty"`
	FieldName  string `json:"fieldName,omitempty"`
	Url        string `json:"url,omitempty"`
}

type AssetUploadTokensResponse struct {
	Tokens []AssetUploadToken `json:"tokens,omitempty"`
}

type AssetUploadTokenOperation struct {
	RecordName string `json:"recordName,omitempty"`
	RecordType string `json:"recordType,omitempty"`
	FieldName  string `json:"fieldName,omitempty"`
	Extension  string `json:"-"`
	UserId     int64  `json:"-"`
}

// Asset upload in a database.
func (s *RecordsService) AssetUpload(ctx context.Context, database Database, req []*AssetUploadTokenOperation) (*AssetUploadTokensResponse, error) {
	var (
		path = "/" + database.String() + s.basePath
		res  AssetUploadTokensResponse
		body = map[string]interface{}{
			"tokens": req,
		}
	)
	if err := s.client.call(ctx, http.MethodPost, path, body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

type UploadRep struct {
	UniqueId   string `json:"-"`
	SingleFile struct {
		// 私有库专用
		// WrappingKey string `json:"wrappingKey"`
		// ReferenceChecksum string `json:"referenceChecksum"`

		FileChecksum string `json:"fileChecksum"`
		Receipt      string `json:"receipt"`
		Size         int64  `json:"size"`
	} `json:"singleFile"`
}

// https://developer.apple.com/library/archive/documentation/DataManagement/Conceptual/CloudKitWebServicesReference/QueryingRecords.html#//apple_ref/doc/uid/TP40015240-CH5-SW4
// query in a database.
func (s *RecordsService) Query(ctx context.Context, database Database, req *QueryFiltersReq, value interface{}) error {
	return s.client.call(ctx, http.MethodPost, "/"+database.String()+s.basePath, req, &value)
}

type SortItem struct {
	FieldName string `json:"fieldName"`
	Ascending bool   `json:"ascending"`
}

type FieldValueItem struct {
	Value interface{} `json:"value"`
}

type QueryFilterItem struct {
	Comparator string          `json:"comparator"`
	FieldName  string          `json:"fieldName"`
	FieldValue *FieldValueItem `json:"fieldValue"`
}

type QueryFiltersItem struct {
	RecordType string             `json:"recordType"`
	Filter     []*QueryFilterItem `json:"filterBy"`
	Sorts      []*SortItem        `json:"sortBy,omitempty"`
}

type QueryFiltersReq struct {
	MaximumResults     int               `json:"maximumResults,omitempty"`
	ResultsLimit       int               `json:"resultsLimit,omitempty"`
	Query              *QueryFiltersItem `json:"query,omitempty"`
	DesiredKeys        []string          `json:"desiredKeys,omitempty"`
	ContinuationMarker string            `json:"continuationMarker,omitempty"`
}
