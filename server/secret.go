package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
)

// resource is the HTTP URL path component for the secrets resource
const resource = "secrets"

// Secret represents a secret from Thycotic Secret Server
type Secret struct {
	Name                                string
	FolderID, ID, SiteID                int
	SecretTemplateID, SecretPolicyID    int
	Active, CheckedOut, CheckOutEnabled bool
	Fields                              []SecretField `json:"Items"`
}

// SecretField is an item (field) in the secret
type SecretField struct {
	ItemID, FieldID, FileAttachmentID                      int
	FieldDescription, FieldName, Filename, ItemValue, Slug string
	IsFile, IsNotes, IsPassword                            bool
}

// Secret gets the secret with id from the Secret Server of the given tenant
func (s Server) Secret(id int) (*Secret, error) {
	return s.getSecret(strconv.Itoa(id))
}

// SecretByPath gets the secret with the given secret path from the Secret
// Server of the given tenant. A secret path includes the secret's fully
// qualified folder path and name, and begins with a leading slash.
func (s Server) SecretByPath(secretPath string) (*Secret, error) {
	return s.getSecret("0?secretPath=" + url.QueryEscape(secretPath))
}

func (s Server) getSecret(secretIdentifier string) (*Secret, error) {
	secret := new(Secret)

	if data, err := s.accessResource("GET", resource, secretIdentifier, nil); err == nil {
		if err = json.Unmarshal(data, secret); err != nil {
			log.Printf("[DEBUG] error parsing response from /%s/%s: %q", resource, secretIdentifier, data)
			return nil, err
		}
	} else {
		return nil, err
	}

	// automatically download file attachments and substitute them for the
	// (dummy) ItemValue, so as to make the process transparent to the caller
	for index, element := range secret.Fields {
		if element.FileAttachmentID != 0 {
			path := fmt.Sprintf("%d/fields/%s", secret.ID, element.Slug)

			if data, err := s.accessResource("GET", resource, path, nil); err == nil {
				secret.Fields[index].ItemValue = string(data)
			} else {
				return nil, err
			}
		}
	}

	return secret, nil
}

// Field returns the value of the field with the name fieldName
func (s Secret) Field(fieldName string) (string, bool) {
	for _, field := range s.Fields {
		if fieldName == field.FieldName || fieldName == field.Slug {
			log.Printf("[DEBUG] field with name '%s' matches '%s'", field.FieldName, fieldName)
			return field.ItemValue, true
		}
	}
	log.Printf("[DEBUG] no matching field for name '%s' in secret '%s'", fieldName, s.Name)
	return "", false
}
