package filesystem

// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"encoding/json"
	"errors"
	"github.com/Azure/go-autorest/autorest"
	"io"
)

// AppendModeType enumerates the values for append mode type.
type AppendModeType string

const (
	// Autocreate specifies the autocreate state for append mode type.
	Autocreate AppendModeType = "autocreate"
)

// Exception enumerates the values for exception.
type Exception string

const (
	// ExceptionAccessControlException specifies the exception access control exception state for exception.
	ExceptionAccessControlException Exception = "AccessControlException"
	// ExceptionBadOffsetException specifies the exception bad offset exception state for exception.
	ExceptionBadOffsetException Exception = "BadOffsetException"
	// ExceptionFileAlreadyExistsException specifies the exception file already exists exception state for exception.
	ExceptionFileAlreadyExistsException Exception = "FileAlreadyExistsException"
	// ExceptionFileNotFoundException specifies the exception file not found exception state for exception.
	ExceptionFileNotFoundException Exception = "FileNotFoundException"
	// ExceptionIllegalArgumentException specifies the exception illegal argument exception state for exception.
	ExceptionIllegalArgumentException Exception = "IllegalArgumentException"
	// ExceptionIOException specifies the exception io exception state for exception.
	ExceptionIOException Exception = "IOException"
	// ExceptionRuntimeException specifies the exception runtime exception state for exception.
	ExceptionRuntimeException Exception = "RuntimeException"
	// ExceptionSecurityException specifies the exception security exception state for exception.
	ExceptionSecurityException Exception = "SecurityException"
	// ExceptionThrottledException specifies the exception throttled exception state for exception.
	ExceptionThrottledException Exception = "ThrottledException"
	// ExceptionUnsupportedOperationException specifies the exception unsupported operation exception state for exception.
	ExceptionUnsupportedOperationException Exception = "UnsupportedOperationException"
)

// ExpiryOptionType enumerates the values for expiry option type.
type ExpiryOptionType string

const (
	// Absolute specifies the absolute state for expiry option type.
	Absolute ExpiryOptionType = "Absolute"
	// NeverExpire specifies the never expire state for expiry option type.
	NeverExpire ExpiryOptionType = "NeverExpire"
	// RelativeToCreationDate specifies the relative to creation date state for expiry option type.
	RelativeToCreationDate ExpiryOptionType = "RelativeToCreationDate"
	// RelativeToNow specifies the relative to now state for expiry option type.
	RelativeToNow ExpiryOptionType = "RelativeToNow"
)

// FileType enumerates the values for file type.
type FileType string

const (
	// DIRECTORY specifies the directory state for file type.
	DIRECTORY FileType = "DIRECTORY"
	// FILE specifies the file state for file type.
	FILE FileType = "FILE"
)

// SyncFlag enumerates the values for sync flag.
type SyncFlag string

const (
	// CLOSE specifies the close state for sync flag.
	CLOSE SyncFlag = "CLOSE"
	// DATA specifies the data state for sync flag.
	DATA SyncFlag = "DATA"
	// METADATA specifies the metadata state for sync flag.
	METADATA SyncFlag = "METADATA"
)

// ACLStatus is data Lake Store file or directory Access Control List information.
type ACLStatus struct {
	Entries    *[]string `json:"entries,omitempty"`
	Group      *string   `json:"group,omitempty"`
	Owner      *string   `json:"owner,omitempty"`
	Permission *string   `json:"permission,omitempty"`
	StickyBit  *bool     `json:"stickyBit,omitempty"`
}

// ACLStatusResult is data Lake Store file or directory Access Control List information.
type ACLStatusResult struct {
	autorest.Response `json:"-"`
	ACLStatus         *ACLStatus `json:"AclStatus,omitempty"`
}

// AdlsAccessControlException is a WebHDFS exception thrown indicating that access is denied due to insufficient
// permissions. Thrown when a 403 error response code is returned (forbidden).
type AdlsAccessControlException struct {
	JavaClassName *string   `json:"javaClassName,omitempty"`
	Message       *string   `json:"message,omitempty"`
	Exception     Exception `json:"exception,omitempty"`
}

// MarshalJSON is the custom marshaler for AdlsAccessControlException.
func (aace AdlsAccessControlException) MarshalJSON() ([]byte, error) {
	aace.Exception = ExceptionAccessControlException
	type Alias AdlsAccessControlException
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(aace),
	})
}

// AsAdlsIllegalArgumentException is the AdlsRemoteException implementation for AdlsAccessControlException.
func (aace AdlsAccessControlException) AsAdlsIllegalArgumentException() (*AdlsIllegalArgumentException, bool) {
	return nil, false
}

// AsAdlsUnsupportedOperationException is the AdlsRemoteException implementation for AdlsAccessControlException.
func (aace AdlsAccessControlException) AsAdlsUnsupportedOperationException() (*AdlsUnsupportedOperationException, bool) {
	return nil, false
}

// AsAdlsSecurityException is the AdlsRemoteException implementation for AdlsAccessControlException.
func (aace AdlsAccessControlException) AsAdlsSecurityException() (*AdlsSecurityException, bool) {
	return nil, false
}

// AsAdlsIOException is the AdlsRemoteException implementation for AdlsAccessControlException.
func (aace AdlsAccessControlException) AsAdlsIOException() (*AdlsIOException, bool) {
	return nil, false
}

// AsAdlsFileNotFoundException is the AdlsRemoteException implementation for AdlsAccessControlException.
func (aace AdlsAccessControlException) AsAdlsFileNotFoundException() (*AdlsFileNotFoundException, bool) {
	return nil, false
}

// AsAdlsFileAlreadyExistsException is the AdlsRemoteException implementation for AdlsAccessControlException.
func (aace AdlsAccessControlException) AsAdlsFileAlreadyExistsException() (*AdlsFileAlreadyExistsException, bool) {
	return nil, false
}

// AsAdlsBadOffsetException is the AdlsRemoteException implementation for AdlsAccessControlException.
func (aace AdlsAccessControlException) AsAdlsBadOffsetException() (*AdlsBadOffsetException, bool) {
	return nil, false
}

// AsAdlsRuntimeException is the AdlsRemoteException implementation for AdlsAccessControlException.
func (aace AdlsAccessControlException) AsAdlsRuntimeException() (*AdlsRuntimeException, bool) {
	return nil, false
}

// AsAdlsAccessControlException is the AdlsRemoteException implementation for AdlsAccessControlException.
func (aace AdlsAccessControlException) AsAdlsAccessControlException() (*AdlsAccessControlException, bool) {
	return &aace, true
}

// AsAdlsThrottledException is the AdlsRemoteException implementation for AdlsAccessControlException.
func (aace AdlsAccessControlException) AsAdlsThrottledException() (*AdlsThrottledException, bool) {
	return nil, false
}

// AdlsBadOffsetException is a WebHDFS exception thrown indicating the append or read is from a bad offset. Thrown when
// a 400 error response code is returned for append and open operations (Bad request).
type AdlsBadOffsetException struct {
	JavaClassName *string   `json:"javaClassName,omitempty"`
	Message       *string   `json:"message,omitempty"`
	Exception     Exception `json:"exception,omitempty"`
}

// MarshalJSON is the custom marshaler for AdlsBadOffsetException.
func (aboe AdlsBadOffsetException) MarshalJSON() ([]byte, error) {
	aboe.Exception = ExceptionBadOffsetException
	type Alias AdlsBadOffsetException
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(aboe),
	})
}

// AsAdlsIllegalArgumentException is the AdlsRemoteException implementation for AdlsBadOffsetException.
func (aboe AdlsBadOffsetException) AsAdlsIllegalArgumentException() (*AdlsIllegalArgumentException, bool) {
	return nil, false
}

// AsAdlsUnsupportedOperationException is the AdlsRemoteException implementation for AdlsBadOffsetException.
func (aboe AdlsBadOffsetException) AsAdlsUnsupportedOperationException() (*AdlsUnsupportedOperationException, bool) {
	return nil, false
}

// AsAdlsSecurityException is the AdlsRemoteException implementation for AdlsBadOffsetException.
func (aboe AdlsBadOffsetException) AsAdlsSecurityException() (*AdlsSecurityException, bool) {
	return nil, false
}

// AsAdlsIOException is the AdlsRemoteException implementation for AdlsBadOffsetException.
func (aboe AdlsBadOffsetException) AsAdlsIOException() (*AdlsIOException, bool) {
	return nil, false
}

// AsAdlsFileNotFoundException is the AdlsRemoteException implementation for AdlsBadOffsetException.
func (aboe AdlsBadOffsetException) AsAdlsFileNotFoundException() (*AdlsFileNotFoundException, bool) {
	return nil, false
}

// AsAdlsFileAlreadyExistsException is the AdlsRemoteException implementation for AdlsBadOffsetException.
func (aboe AdlsBadOffsetException) AsAdlsFileAlreadyExistsException() (*AdlsFileAlreadyExistsException, bool) {
	return nil, false
}

// AsAdlsBadOffsetException is the AdlsRemoteException implementation for AdlsBadOffsetException.
func (aboe AdlsBadOffsetException) AsAdlsBadOffsetException() (*AdlsBadOffsetException, bool) {
	return &aboe, true
}

// AsAdlsRuntimeException is the AdlsRemoteException implementation for AdlsBadOffsetException.
func (aboe AdlsBadOffsetException) AsAdlsRuntimeException() (*AdlsRuntimeException, bool) {
	return nil, false
}

// AsAdlsAccessControlException is the AdlsRemoteException implementation for AdlsBadOffsetException.
func (aboe AdlsBadOffsetException) AsAdlsAccessControlException() (*AdlsAccessControlException, bool) {
	return nil, false
}

// AsAdlsThrottledException is the AdlsRemoteException implementation for AdlsBadOffsetException.
func (aboe AdlsBadOffsetException) AsAdlsThrottledException() (*AdlsThrottledException, bool) {
	return nil, false
}

// AdlsError is data Lake Store filesystem error containing a specific WebHDFS exception.
type AdlsError struct {
	RemoteException AdlsRemoteException `json:"RemoteException,omitempty"`
}

// UnmarshalJSON is the custom unmarshaler for AdlsError struct.
func (ae *AdlsError) UnmarshalJSON(body []byte) error {
	var m map[string]*json.RawMessage
	err := json.Unmarshal(body, &m)
	if err != nil {
		return err
	}
	var v *json.RawMessage

	v = m["RemoteException"]
	if v != nil {
		remoteException, err := unmarshalAdlsRemoteException(*m["RemoteException"])
		if err != nil {
			return err
		}
		ae.RemoteException = remoteException
	}

	return nil
}

// AdlsFileAlreadyExistsException is a WebHDFS exception thrown indicating the file or folder already exists. Thrown
// when a 403 error response code is returned (forbidden).
type AdlsFileAlreadyExistsException struct {
	JavaClassName *string   `json:"javaClassName,omitempty"`
	Message       *string   `json:"message,omitempty"`
	Exception     Exception `json:"exception,omitempty"`
}

// MarshalJSON is the custom marshaler for AdlsFileAlreadyExistsException.
func (afaee AdlsFileAlreadyExistsException) MarshalJSON() ([]byte, error) {
	afaee.Exception = ExceptionFileAlreadyExistsException
	type Alias AdlsFileAlreadyExistsException
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(afaee),
	})
}

// AsAdlsIllegalArgumentException is the AdlsRemoteException implementation for AdlsFileAlreadyExistsException.
func (afaee AdlsFileAlreadyExistsException) AsAdlsIllegalArgumentException() (*AdlsIllegalArgumentException, bool) {
	return nil, false
}

// AsAdlsUnsupportedOperationException is the AdlsRemoteException implementation for AdlsFileAlreadyExistsException.
func (afaee AdlsFileAlreadyExistsException) AsAdlsUnsupportedOperationException() (*AdlsUnsupportedOperationException, bool) {
	return nil, false
}

// AsAdlsSecurityException is the AdlsRemoteException implementation for AdlsFileAlreadyExistsException.
func (afaee AdlsFileAlreadyExistsException) AsAdlsSecurityException() (*AdlsSecurityException, bool) {
	return nil, false
}

// AsAdlsIOException is the AdlsRemoteException implementation for AdlsFileAlreadyExistsException.
func (afaee AdlsFileAlreadyExistsException) AsAdlsIOException() (*AdlsIOException, bool) {
	return nil, false
}

// AsAdlsFileNotFoundException is the AdlsRemoteException implementation for AdlsFileAlreadyExistsException.
func (afaee AdlsFileAlreadyExistsException) AsAdlsFileNotFoundException() (*AdlsFileNotFoundException, bool) {
	return nil, false
}

// AsAdlsFileAlreadyExistsException is the AdlsRemoteException implementation for AdlsFileAlreadyExistsException.
func (afaee AdlsFileAlreadyExistsException) AsAdlsFileAlreadyExistsException() (*AdlsFileAlreadyExistsException, bool) {
	return &afaee, true
}

// AsAdlsBadOffsetException is the AdlsRemoteException implementation for AdlsFileAlreadyExistsException.
func (afaee AdlsFileAlreadyExistsException) AsAdlsBadOffsetException() (*AdlsBadOffsetException, bool) {
	return nil, false
}

// AsAdlsRuntimeException is the AdlsRemoteException implementation for AdlsFileAlreadyExistsException.
func (afaee AdlsFileAlreadyExistsException) AsAdlsRuntimeException() (*AdlsRuntimeException, bool) {
	return nil, false
}

// AsAdlsAccessControlException is the AdlsRemoteException implementation for AdlsFileAlreadyExistsException.
func (afaee AdlsFileAlreadyExistsException) AsAdlsAccessControlException() (*AdlsAccessControlException, bool) {
	return nil, false
}

// AsAdlsThrottledException is the AdlsRemoteException implementation for AdlsFileAlreadyExistsException.
func (afaee AdlsFileAlreadyExistsException) AsAdlsThrottledException() (*AdlsThrottledException, bool) {
	return nil, false
}

// AdlsFileNotFoundException is a WebHDFS exception thrown indicating the file or folder could not be found. Thrown
// when a 404 error response code is returned (not found).
type AdlsFileNotFoundException struct {
	JavaClassName *string   `json:"javaClassName,omitempty"`
	Message       *string   `json:"message,omitempty"`
	Exception     Exception `json:"exception,omitempty"`
}

// MarshalJSON is the custom marshaler for AdlsFileNotFoundException.
func (afnfe AdlsFileNotFoundException) MarshalJSON() ([]byte, error) {
	afnfe.Exception = ExceptionFileNotFoundException
	type Alias AdlsFileNotFoundException
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(afnfe),
	})
}

// AsAdlsIllegalArgumentException is the AdlsRemoteException implementation for AdlsFileNotFoundException.
func (afnfe AdlsFileNotFoundException) AsAdlsIllegalArgumentException() (*AdlsIllegalArgumentException, bool) {
	return nil, false
}

// AsAdlsUnsupportedOperationException is the AdlsRemoteException implementation for AdlsFileNotFoundException.
func (afnfe AdlsFileNotFoundException) AsAdlsUnsupportedOperationException() (*AdlsUnsupportedOperationException, bool) {
	return nil, false
}

// AsAdlsSecurityException is the AdlsRemoteException implementation for AdlsFileNotFoundException.
func (afnfe AdlsFileNotFoundException) AsAdlsSecurityException() (*AdlsSecurityException, bool) {
	return nil, false
}

// AsAdlsIOException is the AdlsRemoteException implementation for AdlsFileNotFoundException.
func (afnfe AdlsFileNotFoundException) AsAdlsIOException() (*AdlsIOException, bool) {
	return nil, false
}

// AsAdlsFileNotFoundException is the AdlsRemoteException implementation for AdlsFileNotFoundException.
func (afnfe AdlsFileNotFoundException) AsAdlsFileNotFoundException() (*AdlsFileNotFoundException, bool) {
	return &afnfe, true
}

// AsAdlsFileAlreadyExistsException is the AdlsRemoteException implementation for AdlsFileNotFoundException.
func (afnfe AdlsFileNotFoundException) AsAdlsFileAlreadyExistsException() (*AdlsFileAlreadyExistsException, bool) {
	return nil, false
}

// AsAdlsBadOffsetException is the AdlsRemoteException implementation for AdlsFileNotFoundException.
func (afnfe AdlsFileNotFoundException) AsAdlsBadOffsetException() (*AdlsBadOffsetException, bool) {
	return nil, false
}

// AsAdlsRuntimeException is the AdlsRemoteException implementation for AdlsFileNotFoundException.
func (afnfe AdlsFileNotFoundException) AsAdlsRuntimeException() (*AdlsRuntimeException, bool) {
	return nil, false
}

// AsAdlsAccessControlException is the AdlsRemoteException implementation for AdlsFileNotFoundException.
func (afnfe AdlsFileNotFoundException) AsAdlsAccessControlException() (*AdlsAccessControlException, bool) {
	return nil, false
}

// AsAdlsThrottledException is the AdlsRemoteException implementation for AdlsFileNotFoundException.
func (afnfe AdlsFileNotFoundException) AsAdlsThrottledException() (*AdlsThrottledException, bool) {
	return nil, false
}

// AdlsIllegalArgumentException is a WebHDFS exception thrown indicating that one more arguments is incorrect. Thrown
// when a 400 error response code is returned (bad request).
type AdlsIllegalArgumentException struct {
	JavaClassName *string   `json:"javaClassName,omitempty"`
	Message       *string   `json:"message,omitempty"`
	Exception     Exception `json:"exception,omitempty"`
}

// MarshalJSON is the custom marshaler for AdlsIllegalArgumentException.
func (aiae AdlsIllegalArgumentException) MarshalJSON() ([]byte, error) {
	aiae.Exception = ExceptionIllegalArgumentException
	type Alias AdlsIllegalArgumentException
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(aiae),
	})
}

// AsAdlsIllegalArgumentException is the AdlsRemoteException implementation for AdlsIllegalArgumentException.
func (aiae AdlsIllegalArgumentException) AsAdlsIllegalArgumentException() (*AdlsIllegalArgumentException, bool) {
	return &aiae, true
}

// AsAdlsUnsupportedOperationException is the AdlsRemoteException implementation for AdlsIllegalArgumentException.
func (aiae AdlsIllegalArgumentException) AsAdlsUnsupportedOperationException() (*AdlsUnsupportedOperationException, bool) {
	return nil, false
}

// AsAdlsSecurityException is the AdlsRemoteException implementation for AdlsIllegalArgumentException.
func (aiae AdlsIllegalArgumentException) AsAdlsSecurityException() (*AdlsSecurityException, bool) {
	return nil, false
}

// AsAdlsIOException is the AdlsRemoteException implementation for AdlsIllegalArgumentException.
func (aiae AdlsIllegalArgumentException) AsAdlsIOException() (*AdlsIOException, bool) {
	return nil, false
}

// AsAdlsFileNotFoundException is the AdlsRemoteException implementation for AdlsIllegalArgumentException.
func (aiae AdlsIllegalArgumentException) AsAdlsFileNotFoundException() (*AdlsFileNotFoundException, bool) {
	return nil, false
}

// AsAdlsFileAlreadyExistsException is the AdlsRemoteException implementation for AdlsIllegalArgumentException.
func (aiae AdlsIllegalArgumentException) AsAdlsFileAlreadyExistsException() (*AdlsFileAlreadyExistsException, bool) {
	return nil, false
}

// AsAdlsBadOffsetException is the AdlsRemoteException implementation for AdlsIllegalArgumentException.
func (aiae AdlsIllegalArgumentException) AsAdlsBadOffsetException() (*AdlsBadOffsetException, bool) {
	return nil, false
}

// AsAdlsRuntimeException is the AdlsRemoteException implementation for AdlsIllegalArgumentException.
func (aiae AdlsIllegalArgumentException) AsAdlsRuntimeException() (*AdlsRuntimeException, bool) {
	return nil, false
}

// AsAdlsAccessControlException is the AdlsRemoteException implementation for AdlsIllegalArgumentException.
func (aiae AdlsIllegalArgumentException) AsAdlsAccessControlException() (*AdlsAccessControlException, bool) {
	return nil, false
}

// AsAdlsThrottledException is the AdlsRemoteException implementation for AdlsIllegalArgumentException.
func (aiae AdlsIllegalArgumentException) AsAdlsThrottledException() (*AdlsThrottledException, bool) {
	return nil, false
}

// AdlsIOException is a WebHDFS exception thrown indicating there was an IO (read or write) error. Thrown when a 403
// error response code is returned (forbidden).
type AdlsIOException struct {
	JavaClassName *string   `json:"javaClassName,omitempty"`
	Message       *string   `json:"message,omitempty"`
	Exception     Exception `json:"exception,omitempty"`
}

// MarshalJSON is the custom marshaler for AdlsIOException.
func (aie AdlsIOException) MarshalJSON() ([]byte, error) {
	aie.Exception = ExceptionIOException
	type Alias AdlsIOException
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(aie),
	})
}

// AsAdlsIllegalArgumentException is the AdlsRemoteException implementation for AdlsIOException.
func (aie AdlsIOException) AsAdlsIllegalArgumentException() (*AdlsIllegalArgumentException, bool) {
	return nil, false
}

// AsAdlsUnsupportedOperationException is the AdlsRemoteException implementation for AdlsIOException.
func (aie AdlsIOException) AsAdlsUnsupportedOperationException() (*AdlsUnsupportedOperationException, bool) {
	return nil, false
}

// AsAdlsSecurityException is the AdlsRemoteException implementation for AdlsIOException.
func (aie AdlsIOException) AsAdlsSecurityException() (*AdlsSecurityException, bool) {
	return nil, false
}

// AsAdlsIOException is the AdlsRemoteException implementation for AdlsIOException.
func (aie AdlsIOException) AsAdlsIOException() (*AdlsIOException, bool) {
	return &aie, true
}

// AsAdlsFileNotFoundException is the AdlsRemoteException implementation for AdlsIOException.
func (aie AdlsIOException) AsAdlsFileNotFoundException() (*AdlsFileNotFoundException, bool) {
	return nil, false
}

// AsAdlsFileAlreadyExistsException is the AdlsRemoteException implementation for AdlsIOException.
func (aie AdlsIOException) AsAdlsFileAlreadyExistsException() (*AdlsFileAlreadyExistsException, bool) {
	return nil, false
}

// AsAdlsBadOffsetException is the AdlsRemoteException implementation for AdlsIOException.
func (aie AdlsIOException) AsAdlsBadOffsetException() (*AdlsBadOffsetException, bool) {
	return nil, false
}

// AsAdlsRuntimeException is the AdlsRemoteException implementation for AdlsIOException.
func (aie AdlsIOException) AsAdlsRuntimeException() (*AdlsRuntimeException, bool) {
	return nil, false
}

// AsAdlsAccessControlException is the AdlsRemoteException implementation for AdlsIOException.
func (aie AdlsIOException) AsAdlsAccessControlException() (*AdlsAccessControlException, bool) {
	return nil, false
}

// AsAdlsThrottledException is the AdlsRemoteException implementation for AdlsIOException.
func (aie AdlsIOException) AsAdlsThrottledException() (*AdlsThrottledException, bool) {
	return nil, false
}

// AdlsRemoteException is data Lake Store filesystem exception based on the WebHDFS definition for RemoteExceptions.
// This is a WebHDFS 'catch all' exception
type AdlsRemoteException interface {
	AsAdlsIllegalArgumentException() (*AdlsIllegalArgumentException, bool)
	AsAdlsUnsupportedOperationException() (*AdlsUnsupportedOperationException, bool)
	AsAdlsSecurityException() (*AdlsSecurityException, bool)
	AsAdlsIOException() (*AdlsIOException, bool)
	AsAdlsFileNotFoundException() (*AdlsFileNotFoundException, bool)
	AsAdlsFileAlreadyExistsException() (*AdlsFileAlreadyExistsException, bool)
	AsAdlsBadOffsetException() (*AdlsBadOffsetException, bool)
	AsAdlsRuntimeException() (*AdlsRuntimeException, bool)
	AsAdlsAccessControlException() (*AdlsAccessControlException, bool)
	AsAdlsThrottledException() (*AdlsThrottledException, bool)
}

func unmarshalAdlsRemoteException(body []byte) (AdlsRemoteException, error) {
	var m map[string]interface{}
	err := json.Unmarshal(body, &m)
	if err != nil {
		return nil, err
	}

	switch m["exception"] {
	case string(ExceptionIllegalArgumentException):
		var aiae AdlsIllegalArgumentException
		err := json.Unmarshal(body, &aiae)
		return aiae, err
	case string(ExceptionUnsupportedOperationException):
		var auoe AdlsUnsupportedOperationException
		err := json.Unmarshal(body, &auoe)
		return auoe, err
	case string(ExceptionSecurityException):
		var ase AdlsSecurityException
		err := json.Unmarshal(body, &ase)
		return ase, err
	case string(ExceptionIOException):
		var aie AdlsIOException
		err := json.Unmarshal(body, &aie)
		return aie, err
	case string(ExceptionFileNotFoundException):
		var afnfe AdlsFileNotFoundException
		err := json.Unmarshal(body, &afnfe)
		return afnfe, err
	case string(ExceptionFileAlreadyExistsException):
		var afaee AdlsFileAlreadyExistsException
		err := json.Unmarshal(body, &afaee)
		return afaee, err
	case string(ExceptionBadOffsetException):
		var aboe AdlsBadOffsetException
		err := json.Unmarshal(body, &aboe)
		return aboe, err
	case string(ExceptionRuntimeException):
		var are AdlsRuntimeException
		err := json.Unmarshal(body, &are)
		return are, err
	case string(ExceptionAccessControlException):
		var aace AdlsAccessControlException
		err := json.Unmarshal(body, &aace)
		return aace, err
	case string(ExceptionThrottledException):
		var ate AdlsThrottledException
		err := json.Unmarshal(body, &ate)
		return ate, err
	default:
		return nil, errors.New("Unsupported type")
	}
}
func unmarshalAdlsRemoteExceptionArray(body []byte) ([]AdlsRemoteException, error) {
	var rawMessages []*json.RawMessage
	err := json.Unmarshal(body, &rawMessages)
	if err != nil {
		return nil, err
	}

	areArray := make([]AdlsRemoteException, len(rawMessages))

	for index, rawMessage := range rawMessages {
		are, err := unmarshalAdlsRemoteException(*rawMessage)
		if err != nil {
			return nil, err
		}
		areArray[index] = are
	}
	return areArray, nil
}

// AdlsRuntimeException is a WebHDFS exception thrown when an unexpected error occurs during an operation. Thrown when
// a 500 error response code is returned (Internal server error).
type AdlsRuntimeException struct {
	JavaClassName *string   `json:"javaClassName,omitempty"`
	Message       *string   `json:"message,omitempty"`
	Exception     Exception `json:"exception,omitempty"`
}

// MarshalJSON is the custom marshaler for AdlsRuntimeException.
func (are AdlsRuntimeException) MarshalJSON() ([]byte, error) {
	are.Exception = ExceptionRuntimeException
	type Alias AdlsRuntimeException
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(are),
	})
}

// AsAdlsIllegalArgumentException is the AdlsRemoteException implementation for AdlsRuntimeException.
func (are AdlsRuntimeException) AsAdlsIllegalArgumentException() (*AdlsIllegalArgumentException, bool) {
	return nil, false
}

// AsAdlsUnsupportedOperationException is the AdlsRemoteException implementation for AdlsRuntimeException.
func (are AdlsRuntimeException) AsAdlsUnsupportedOperationException() (*AdlsUnsupportedOperationException, bool) {
	return nil, false
}

// AsAdlsSecurityException is the AdlsRemoteException implementation for AdlsRuntimeException.
func (are AdlsRuntimeException) AsAdlsSecurityException() (*AdlsSecurityException, bool) {
	return nil, false
}

// AsAdlsIOException is the AdlsRemoteException implementation for AdlsRuntimeException.
func (are AdlsRuntimeException) AsAdlsIOException() (*AdlsIOException, bool) {
	return nil, false
}

// AsAdlsFileNotFoundException is the AdlsRemoteException implementation for AdlsRuntimeException.
func (are AdlsRuntimeException) AsAdlsFileNotFoundException() (*AdlsFileNotFoundException, bool) {
	return nil, false
}

// AsAdlsFileAlreadyExistsException is the AdlsRemoteException implementation for AdlsRuntimeException.
func (are AdlsRuntimeException) AsAdlsFileAlreadyExistsException() (*AdlsFileAlreadyExistsException, bool) {
	return nil, false
}

// AsAdlsBadOffsetException is the AdlsRemoteException implementation for AdlsRuntimeException.
func (are AdlsRuntimeException) AsAdlsBadOffsetException() (*AdlsBadOffsetException, bool) {
	return nil, false
}

// AsAdlsRuntimeException is the AdlsRemoteException implementation for AdlsRuntimeException.
func (are AdlsRuntimeException) AsAdlsRuntimeException() (*AdlsRuntimeException, bool) {
	return &are, true
}

// AsAdlsAccessControlException is the AdlsRemoteException implementation for AdlsRuntimeException.
func (are AdlsRuntimeException) AsAdlsAccessControlException() (*AdlsAccessControlException, bool) {
	return nil, false
}

// AsAdlsThrottledException is the AdlsRemoteException implementation for AdlsRuntimeException.
func (are AdlsRuntimeException) AsAdlsThrottledException() (*AdlsThrottledException, bool) {
	return nil, false
}

// AdlsSecurityException is a WebHDFS exception thrown indicating that access is denied. Thrown when a 401 error
// response code is returned (Unauthorized).
type AdlsSecurityException struct {
	JavaClassName *string   `json:"javaClassName,omitempty"`
	Message       *string   `json:"message,omitempty"`
	Exception     Exception `json:"exception,omitempty"`
}

// MarshalJSON is the custom marshaler for AdlsSecurityException.
func (ase AdlsSecurityException) MarshalJSON() ([]byte, error) {
	ase.Exception = ExceptionSecurityException
	type Alias AdlsSecurityException
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(ase),
	})
}

// AsAdlsIllegalArgumentException is the AdlsRemoteException implementation for AdlsSecurityException.
func (ase AdlsSecurityException) AsAdlsIllegalArgumentException() (*AdlsIllegalArgumentException, bool) {
	return nil, false
}

// AsAdlsUnsupportedOperationException is the AdlsRemoteException implementation for AdlsSecurityException.
func (ase AdlsSecurityException) AsAdlsUnsupportedOperationException() (*AdlsUnsupportedOperationException, bool) {
	return nil, false
}

// AsAdlsSecurityException is the AdlsRemoteException implementation for AdlsSecurityException.
func (ase AdlsSecurityException) AsAdlsSecurityException() (*AdlsSecurityException, bool) {
	return &ase, true
}

// AsAdlsIOException is the AdlsRemoteException implementation for AdlsSecurityException.
func (ase AdlsSecurityException) AsAdlsIOException() (*AdlsIOException, bool) {
	return nil, false
}

// AsAdlsFileNotFoundException is the AdlsRemoteException implementation for AdlsSecurityException.
func (ase AdlsSecurityException) AsAdlsFileNotFoundException() (*AdlsFileNotFoundException, bool) {
	return nil, false
}

// AsAdlsFileAlreadyExistsException is the AdlsRemoteException implementation for AdlsSecurityException.
func (ase AdlsSecurityException) AsAdlsFileAlreadyExistsException() (*AdlsFileAlreadyExistsException, bool) {
	return nil, false
}

// AsAdlsBadOffsetException is the AdlsRemoteException implementation for AdlsSecurityException.
func (ase AdlsSecurityException) AsAdlsBadOffsetException() (*AdlsBadOffsetException, bool) {
	return nil, false
}

// AsAdlsRuntimeException is the AdlsRemoteException implementation for AdlsSecurityException.
func (ase AdlsSecurityException) AsAdlsRuntimeException() (*AdlsRuntimeException, bool) {
	return nil, false
}

// AsAdlsAccessControlException is the AdlsRemoteException implementation for AdlsSecurityException.
func (ase AdlsSecurityException) AsAdlsAccessControlException() (*AdlsAccessControlException, bool) {
	return nil, false
}

// AsAdlsThrottledException is the AdlsRemoteException implementation for AdlsSecurityException.
func (ase AdlsSecurityException) AsAdlsThrottledException() (*AdlsThrottledException, bool) {
	return nil, false
}

// AdlsThrottledException is a WebHDFS exception thrown indicating that the request is being throttled. Reducing the
// number of requests or request size helps to mitigate this error.
type AdlsThrottledException struct {
	JavaClassName *string   `json:"javaClassName,omitempty"`
	Message       *string   `json:"message,omitempty"`
	Exception     Exception `json:"exception,omitempty"`
}

// MarshalJSON is the custom marshaler for AdlsThrottledException.
func (ate AdlsThrottledException) MarshalJSON() ([]byte, error) {
	ate.Exception = ExceptionThrottledException
	type Alias AdlsThrottledException
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(ate),
	})
}

// AsAdlsIllegalArgumentException is the AdlsRemoteException implementation for AdlsThrottledException.
func (ate AdlsThrottledException) AsAdlsIllegalArgumentException() (*AdlsIllegalArgumentException, bool) {
	return nil, false
}

// AsAdlsUnsupportedOperationException is the AdlsRemoteException implementation for AdlsThrottledException.
func (ate AdlsThrottledException) AsAdlsUnsupportedOperationException() (*AdlsUnsupportedOperationException, bool) {
	return nil, false
}

// AsAdlsSecurityException is the AdlsRemoteException implementation for AdlsThrottledException.
func (ate AdlsThrottledException) AsAdlsSecurityException() (*AdlsSecurityException, bool) {
	return nil, false
}

// AsAdlsIOException is the AdlsRemoteException implementation for AdlsThrottledException.
func (ate AdlsThrottledException) AsAdlsIOException() (*AdlsIOException, bool) {
	return nil, false
}

// AsAdlsFileNotFoundException is the AdlsRemoteException implementation for AdlsThrottledException.
func (ate AdlsThrottledException) AsAdlsFileNotFoundException() (*AdlsFileNotFoundException, bool) {
	return nil, false
}

// AsAdlsFileAlreadyExistsException is the AdlsRemoteException implementation for AdlsThrottledException.
func (ate AdlsThrottledException) AsAdlsFileAlreadyExistsException() (*AdlsFileAlreadyExistsException, bool) {
	return nil, false
}

// AsAdlsBadOffsetException is the AdlsRemoteException implementation for AdlsThrottledException.
func (ate AdlsThrottledException) AsAdlsBadOffsetException() (*AdlsBadOffsetException, bool) {
	return nil, false
}

// AsAdlsRuntimeException is the AdlsRemoteException implementation for AdlsThrottledException.
func (ate AdlsThrottledException) AsAdlsRuntimeException() (*AdlsRuntimeException, bool) {
	return nil, false
}

// AsAdlsAccessControlException is the AdlsRemoteException implementation for AdlsThrottledException.
func (ate AdlsThrottledException) AsAdlsAccessControlException() (*AdlsAccessControlException, bool) {
	return nil, false
}

// AsAdlsThrottledException is the AdlsRemoteException implementation for AdlsThrottledException.
func (ate AdlsThrottledException) AsAdlsThrottledException() (*AdlsThrottledException, bool) {
	return &ate, true
}

// AdlsUnsupportedOperationException is a WebHDFS exception thrown indicating that the requested operation is not
// supported. Thrown when a 400 error response code is returned (bad request).
type AdlsUnsupportedOperationException struct {
	JavaClassName *string   `json:"javaClassName,omitempty"`
	Message       *string   `json:"message,omitempty"`
	Exception     Exception `json:"exception,omitempty"`
}

// MarshalJSON is the custom marshaler for AdlsUnsupportedOperationException.
func (auoe AdlsUnsupportedOperationException) MarshalJSON() ([]byte, error) {
	auoe.Exception = ExceptionUnsupportedOperationException
	type Alias AdlsUnsupportedOperationException
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(auoe),
	})
}

// AsAdlsIllegalArgumentException is the AdlsRemoteException implementation for AdlsUnsupportedOperationException.
func (auoe AdlsUnsupportedOperationException) AsAdlsIllegalArgumentException() (*AdlsIllegalArgumentException, bool) {
	return nil, false
}

// AsAdlsUnsupportedOperationException is the AdlsRemoteException implementation for AdlsUnsupportedOperationException.
func (auoe AdlsUnsupportedOperationException) AsAdlsUnsupportedOperationException() (*AdlsUnsupportedOperationException, bool) {
	return &auoe, true
}

// AsAdlsSecurityException is the AdlsRemoteException implementation for AdlsUnsupportedOperationException.
func (auoe AdlsUnsupportedOperationException) AsAdlsSecurityException() (*AdlsSecurityException, bool) {
	return nil, false
}

// AsAdlsIOException is the AdlsRemoteException implementation for AdlsUnsupportedOperationException.
func (auoe AdlsUnsupportedOperationException) AsAdlsIOException() (*AdlsIOException, bool) {
	return nil, false
}

// AsAdlsFileNotFoundException is the AdlsRemoteException implementation for AdlsUnsupportedOperationException.
func (auoe AdlsUnsupportedOperationException) AsAdlsFileNotFoundException() (*AdlsFileNotFoundException, bool) {
	return nil, false
}

// AsAdlsFileAlreadyExistsException is the AdlsRemoteException implementation for AdlsUnsupportedOperationException.
func (auoe AdlsUnsupportedOperationException) AsAdlsFileAlreadyExistsException() (*AdlsFileAlreadyExistsException, bool) {
	return nil, false
}

// AsAdlsBadOffsetException is the AdlsRemoteException implementation for AdlsUnsupportedOperationException.
func (auoe AdlsUnsupportedOperationException) AsAdlsBadOffsetException() (*AdlsBadOffsetException, bool) {
	return nil, false
}

// AsAdlsRuntimeException is the AdlsRemoteException implementation for AdlsUnsupportedOperationException.
func (auoe AdlsUnsupportedOperationException) AsAdlsRuntimeException() (*AdlsRuntimeException, bool) {
	return nil, false
}

// AsAdlsAccessControlException is the AdlsRemoteException implementation for AdlsUnsupportedOperationException.
func (auoe AdlsUnsupportedOperationException) AsAdlsAccessControlException() (*AdlsAccessControlException, bool) {
	return nil, false
}

// AsAdlsThrottledException is the AdlsRemoteException implementation for AdlsUnsupportedOperationException.
func (auoe AdlsUnsupportedOperationException) AsAdlsThrottledException() (*AdlsThrottledException, bool) {
	return nil, false
}

// ContentSummary is data Lake Store content summary information
type ContentSummary struct {
	DirectoryCount *int64 `json:"directoryCount,omitempty"`
	FileCount      *int64 `json:"fileCount,omitempty"`
	Length         *int64 `json:"length,omitempty"`
	SpaceConsumed  *int64 `json:"spaceConsumed,omitempty"`
}

// ContentSummaryResult is data Lake Store filesystem content summary information response.
type ContentSummaryResult struct {
	autorest.Response `json:"-"`
	ContentSummary    *ContentSummary `json:"ContentSummary,omitempty"`
}

// FileOperationResult is the result of the request or operation.
type FileOperationResult struct {
	autorest.Response `json:"-"`
	OperationResult   *bool `json:"boolean,omitempty"`
}

// FileStatuses is data Lake Store file status list information.
type FileStatuses struct {
	FileStatus *[]FileStatusProperties `json:"FileStatus,omitempty"`
}

// FileStatusesResult is data Lake Store filesystem file status list information response.
type FileStatusesResult struct {
	autorest.Response `json:"-"`
	FileStatuses      *FileStatuses `json:"FileStatuses,omitempty"`
}

// FileStatusProperties is data Lake Store file or directory information.
type FileStatusProperties struct {
	AccessTime       *int64   `json:"accessTime,omitempty"`
	BlockSize        *int64   `json:"blockSize,omitempty"`
	ChildrenNum      *int64   `json:"childrenNum,omitempty"`
	ExpirationTime   *int64   `json:"msExpirationTime,omitempty"`
	Group            *string  `json:"group,omitempty"`
	Length           *int64   `json:"length,omitempty"`
	ModificationTime *int64   `json:"modificationTime,omitempty"`
	Owner            *string  `json:"owner,omitempty"`
	PathSuffix       *string  `json:"pathSuffix,omitempty"`
	Permission       *string  `json:"permission,omitempty"`
	Type             FileType `json:"type,omitempty"`
	ACLBit           *bool    `json:"aclBit,omitempty"`
}

// FileStatusResult is data Lake Store filesystem file status information response.
type FileStatusResult struct {
	autorest.Response `json:"-"`
	FileStatus        *FileStatusProperties `json:"FileStatus,omitempty"`
}

// ReadCloser is
type ReadCloser struct {
	autorest.Response `json:"-"`
	Value             *io.ReadCloser `json:"value,omitempty"`
}
