package server

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
)

// TestSecret tests Secret. Referred to as "Test #1" in the README.
func TestSecret(t *testing.T) {
	tss, err := initServer()
	if err != nil {
		t.Error("configuring the Server:", err)
		return
	}

	id := initIntegerFromEnv("TSS_SECRET_ID", t)
	if id < 0 {
		return
	}

	s, err := tss.Secret(id)

	if err != nil {
		t.Error("calling server.Secret:", err)
		return
	}

	if s == nil {
		t.Error("secret data is nil")
	}

	if _, ok := s.Field("password"); !ok {
		t.Error("no password field")
	}

	if _, ok := s.Field("nonexistent"); ok {
		t.Error("s.Field says nonexistent field exists")
	}
}

// TestSecretCRUD tests the creation, read, update, and delete of a Secret.
// Referred to as "Test #2" in the README.
func TestSecretCRUD(t *testing.T) {

	// Initialize
	tss, err := initServer()
	if err != nil {
		t.Error("configuring the Server:", err)
		return
	}
	siteId := initIntegerFromEnv("TSS_SITE_ID", t)
	folderId := initIntegerFromEnv("TSS_FOLDER_ID", t)
	templateId := initIntegerFromEnv("TSS_TEMPLATE_ID", t)
	fieldId := initIntegerFromEnv("TSS_FIELD_ID", t)
	if siteId < 0 || folderId < 0 || templateId < 0 || fieldId < 0 {
		return
	}

	// Test creation of a new secret
	refSecret := new(Secret)
	password := "Shhhhhhhhhhh!123"
	refSecret.Name = "Test Secret"
	refSecret.SiteID = siteId
	refSecret.FolderID = folderId
	refSecret.SecretTemplateID = templateId
	refSecret.Fields = make([]SecretField, 1)
	refSecret.Fields[0].FieldID = fieldId
	refSecret.Fields[0].ItemValue = password
	sc, err := tss.CreateSecret(*refSecret)
	if err != nil { t.Error("calling server.CreateSecret:", err); return }
	if sc == nil { t.Error("created secret data is nil"); return }
	if !validate("created secret folder id", folderId, sc.FolderID, t) { return }
	if !validate("created secret template id", templateId, sc.SecretTemplateID, t) { return }
	if !validate("created secret site id", siteId, sc.SiteID, t) { return }
	createdPassword, matched := sc.FieldById(fieldId)
	if !matched { t.Errorf("created secret does not have a password field with the given field id '%d':", fieldId); return }
	if !validate("created secret password value", password, createdPassword, t) { return }

	// Test the read of the new secret
	sr, err := tss.Secret(sc.ID)
	if err != nil { t.Error("calling server.Secret:", err); return }
	if sr == nil { t.Error("read secret data is nil"); return }
	if !validate("read secret folder id", folderId, sr.FolderID, t) { return }
	if !validate("read secret template id", templateId, sr.SecretTemplateID, t) { return }
	if !validate("read secret site id", siteId, sr.SiteID, t) { return }
	readPassword, matched := sr.FieldById(fieldId)
	if !matched { t.Errorf("read secret does not have a password field with the given field id '%d':", fieldId); return }
	if !validate("read secret password value", password, readPassword, t) { return }

	// Test the update of the new secret
	newPassword := password + "updated"
	refSecret.ID = sc.ID
	refSecret.Fields[0].ItemValue = newPassword
	su, err := tss.UpdateSecret(*refSecret)
	if err != nil { t.Error("calling server.UpdateSecret:", err); return }
	if su == nil { t.Error("updated secret data is nil"); return }
	if !validate("updated secret folder id", folderId, su.FolderID, t) { return }
	if !validate("updated secret template id", templateId, su.SecretTemplateID, t) { return }
	if !validate("updated secret site id", siteId, su.SiteID, t) { return }
	updatedPassword, matched := su.FieldById(fieldId)
	if !matched { t.Errorf("updated secret does not have a password field with the given field id '%d':", fieldId); return }
	if !validate("updated secret password value", newPassword, updatedPassword, t) { return }

	// Test the deletion of the new secret
	err = tss.DeleteSecret(sc.ID)
	if err != nil { t.Error("calling server.DeleteSecret:", err); return }

	// Test read of the deleted secret fails
	s, err := tss.Secret(sc.ID)
	if s != nil { t.Errorf("deleted secret with id '%d' returned from read", sc.ID) }
}

// TestSecretCRUDForSSHTemplate tests the creation, read, update, and delete
// of a Secret which uses an SSH key template, that is, a template with extended
// mappings that support SSH keys. Referred to as "Test #3" in the README.
func TestSecretCRUDForSSHTemplate(t *testing.T) {

	// Initialize
	tss, err := initServer()
	if err != nil {
		t.Error("configuring the Server:", err)
		return
	}
	siteId := initIntegerFromEnv("TSS_SITE_ID", t)
	folderId := initIntegerFromEnv("TSS_FOLDER_ID", t)
	templateId := initIntegerFromEnv("TSS_SSH_KEY_TEMPLATE_ID", t)
	publicKeyFieldId := initIntegerFromEnv("TSS_SSH_PUBLIC_FIELD_ID", t)
	privateKeyFieldId := initIntegerFromEnv("TSS_SSH_PRIVATE_FIELD_ID", t)
	passphraseFieldId := initIntegerFromEnv("TSS_SSH_PASSPHRASE_FIELD_ID", t)
	if siteId < 0 || folderId < 0 || templateId < 0 || publicKeyFieldId < 0 || privateKeyFieldId < 0 || passphraseFieldId < 0 {
		return
	}

	// Test creation of a new secret
	refSecret := new(Secret)
	refSecret.Name = "Test SSH Key Secret"
	refSecret.SiteID = siteId
	refSecret.FolderID = folderId
	refSecret.SecretTemplateID = templateId
	refSecret.SshKeyArgs = &SshKeyArgs{}
	refSecret.SshKeyArgs.GenerateSshKeys = true
	refSecret.SshKeyArgs.GeneratePassphrase = true
	refSecret.Fields = make([]SecretField, 2)
	refSecret.Fields[0].FieldID = publicKeyFieldId
	refSecret.Fields[0].Filename = "" // Let the server generate the name
	refSecret.Fields[1].FieldID = privateKeyFieldId
	refSecret.Fields[1].Filename = "My Private Key.pem"
	sc, err := tss.CreateSecret(*refSecret)
	if err != nil { t.Error("calling server.CreateSecret:", err); return }
	if sc == nil { t.Error("created secret data is nil"); return }
	if !validate("created secret name", "Test SSH Key Secret", sc.Name, t) { return }
	if !validate("created secret folder id", folderId, sc.FolderID, t) { return }
	if !validate("created secret template id", templateId, sc.SecretTemplateID, t) { return }
	if !validate("created secret site id", siteId, sc.SiteID, t) { return }
	if publicKeyField := getField(sc, publicKeyFieldId); publicKeyField != nil {
		if !validate("created secret public key field is a file field", true, publicKeyField.IsFile, t) { return }
		if !validate("created secret public key field has a generated value", true, len(publicKeyField.ItemValue) > 100, t) { return }
		if !validate("created secret public key field has a generated file name", publicKeyField.FieldName, publicKeyField.Filename, t) { return }
	} else {
		t.Errorf("created secret does not have a public key field at the given field id '%d':", publicKeyFieldId)
		return
	}
	if privateKeyField := getField(sc, privateKeyFieldId); privateKeyField != nil {
		if !validate("created secret private key field is a file field", true, privateKeyField.IsFile, t) { return }
		if !validate("created secret private key field has a generated value", true, len(privateKeyField.ItemValue) > 100, t) { return }
		if !validate("created secret private key field has the given file name", "My Private Key.pem", privateKeyField.Filename, t) { return }
	} else {
		t.Errorf("created secret does not have a private key field at the given field id '%d':", privateKeyFieldId)
		return
	}
	if passphraseField := getField(sc, passphraseFieldId); passphraseField != nil {
		if !validate("created secret passphrase field is a password field", true, passphraseField.IsPassword, t) { return }
		if !validate("created secret passphrase field has a value", true, len(passphraseField.ItemValue) > 10, t) { return }
	} else {
		t.Errorf("created secret does not have a passphrase field at the given field id '%d':", passphraseFieldId)
		return
	}

	// Test the read of the new secret
	sr, err := tss.Secret(sc.ID)
	if err != nil { t.Error("calling server.Secret:", err); return }
	if sr == nil { t.Error("read secret data is nil"); return }
	if !validate("read secret name", "Test SSH Key Secret", sr.Name, t) { return }
	if !validate("read secret folder id", folderId, sr.FolderID, t) { return }
	if !validate("read secret template id", templateId, sr.SecretTemplateID, t) { return }
	if !validate("read secret site id", siteId, sr.SiteID, t) { return }
	if publicKeyField := getField(sr, publicKeyFieldId); publicKeyField != nil {
		if !validate("read secret public key field is a file field", true, publicKeyField.IsFile, t) { return }
		if !validate("read secret public key field has a value", true, len(publicKeyField.ItemValue) > 100, t) { return }
		if !validate("read secret public key field has a generated file name", publicKeyField.FieldName, publicKeyField.Filename, t) { return }
	} else {
		t.Errorf("read secret does not have a public key field at the given field id '%d':", publicKeyFieldId)
		return
	}
	if privateKeyField := getField(sr, privateKeyFieldId); privateKeyField != nil {
		if !validate("read secret private key field is a file field", true, privateKeyField.IsFile, t) { return }
		if !validate("read secret private key field has a value", true, len(privateKeyField.ItemValue) > 100, t) { return }
		if !validate("read secret private key field has the given file name", "My Private Key.pem", privateKeyField.Filename, t) { return }
	} else {
		t.Errorf("read secret does not have a private key field at the given field id '%d':", privateKeyFieldId)
		return
	}
	if passphraseField := getField(sr, passphraseFieldId); passphraseField != nil {
		if !validate("read secret passphrase field is a password field", true, passphraseField.IsPassword, t) { return }
		if !validate("read secret passphrase field has a value", true, len(passphraseField.ItemValue) > 10, t) { return }
	} else {
		t.Errorf("read secret does not have a passphrase field at the given field id '%d':", passphraseFieldId)
		return
	}

	// Test the update of the new secret
	sc.Name = sc.Name + " (Updated)"
	sc.SshKeyArgs = nil
	sc.Fields[0].Filename = "New Filename.txt"
	su, err := tss.UpdateSecret(*sc)
	if err != nil { t.Error("calling server.UpdateSecret:", err); return }
	if su == nil { t.Error("updated secret data is nil"); return }
	if !validate("updated secret name", "Test SSH Key Secret (Updated)", su.Name, t) { return }
	if !validate("updated secret folder id", folderId, su.FolderID, t) { return }
	if !validate("updated secret template id", templateId, su.SecretTemplateID, t) { return }
	if !validate("updated secret site id", siteId, su.SiteID, t) { return }
	if publicKeyField := getField(su, publicKeyFieldId); publicKeyField != nil {
		if !validate("updated secret public key field is a file field", true, publicKeyField.IsFile, t) { return }
		if !validate("updated secret public key field has a value", true, len(publicKeyField.ItemValue) > 100, t) { return }
		if !validate("updated secret public key field has the given file name", "New Filename.txt", publicKeyField.Filename, t) { return }
	} else {
		t.Errorf("updated secret does not have a public key field at the given field id '%d':", publicKeyFieldId)
		return
	}
	if privateKeyField := getField(su, privateKeyFieldId); privateKeyField != nil {
		if !validate("updated secret private key field is a file field", true, privateKeyField.IsFile, t) { return }
		if !validate("updated secret private key field has a value", true, len(privateKeyField.ItemValue) > 100, t) { return }
		if !validate("updated secret private key field has the given file name", "My Private Key.pem", privateKeyField.Filename, t) { return }
	} else {
		t.Errorf("updated secret does not have a private key field at the given field id '%d':", privateKeyFieldId)
		return
	}
	if passphraseField := getField(su, passphraseFieldId); passphraseField != nil {
		if !validate("updated secret passphrase field is a password field", true, passphraseField.IsPassword, t) { return }
		if !validate("updated secret passphrase field has a value", true, len(passphraseField.ItemValue) > 10, t) { return }
	} else {
		t.Errorf("updated secret does not have a passphrase field at the given field id '%d':", passphraseFieldId)
		return
	}

	// Test the deletion of the new secret
	err = tss.DeleteSecret(sc.ID)
	if err != nil { t.Error("calling server.DeleteSecret:", err); return }

	// Test read of the deleted secret fails
	s, err := tss.Secret(sc.ID)
	if s != nil { t.Errorf("deleted secret with id '%d' returned from read", sc.ID) }
}

func initServer() (*Server, error) {
	var config *Configuration

	if cj, err := ioutil.ReadFile("../test_config.json"); err == nil {
		config = new(Configuration)

		json.Unmarshal(cj, &config)
	} else {
		config = &Configuration{
			Credentials: UserCredential{
				Username: os.Getenv("TSS_USERNAME"),
				Password: os.Getenv("TSS_PASSWORD"),
			},
			// Expecting either the tenant or URL to be set
			Tenant:    os.Getenv("TSS_TENANT"),
			ServerURL: os.Getenv("TSS_SERVER_URL"),
		}
	}
	return New(*config)
}

// initIntegerFromEnv reads the given environment variable and if it's declared, parses it to an integer. Otherwise,
// returns a default integer of '1'.
func initIntegerFromEnv(envVarName string, t *testing.T) int {
	intValue := 1
	valueFromEnv := os.Getenv(envVarName)
	if valueFromEnv != "" {
		var err error
		intValue, err = strconv.Atoi(valueFromEnv)
		if err != nil {
			t.Errorf("%s must be an integer: %s", envVarName, err)
			return -1
		}
	}
	return intValue
}

func validate(label string, expected interface{}, found interface{}, t *testing.T) bool {
	if expected != found {
		t.Errorf("expecting '%s' to be '%q', but found '%q' instead.", label, expected, found)
		return false
	}
	return true
}

func getField(secret *Secret, fieldId int) *SecretField {
	for _, field := range secret.Fields {
		if field.FieldID == fieldId { return &field }
	}
	return nil
}