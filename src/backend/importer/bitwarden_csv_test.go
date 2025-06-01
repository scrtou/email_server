package importer

import (
	"reflect"
	"strings"
	"testing"

	"email_server/models"
)

func TestParseBitwardenCSV(t *testing.T) {
	csvData := "folder,favorite,type,name,notes,fields,reprompt,login_uri,login_username,login_password,login_totp\n" +
		`Work,0,login,Example Site,"Some notes here","{""custom_field_1"":""value1"",""secret_code"":""1234""}",0,https://example.com,user@example.com,password123,otpauth://totp/Example:user@example.com?secret=JBSWY3DPEHPK3PXP&issuer=Example` + "\n" +
		`,1,login,Another Site,,,1,https://another.com,anotheruser,anotherpass,` + "\n" +
		`,,login,Site with TOTP,,,,https://withtotp.com,totpuser,totppass,otpauth://totp/WithTOTP:totpuser?secret=ABCDEFGH&issuer=WithTOTP` + "\n" +
		`Empty Folder,,,Empty Name,,,,"",empty_user,empty_pass,` + "\n" +
		`"Folder, with comma",0,login,"Name, with comma","Note with ""quotes""",,"0","https://comma.com","comma,user","comma,pass",` + "\n" +
		`,,,,,,,,,,` + "\n"

	t.Run("Basic Parsing with Passwords", func(t *testing.T) {
		reader := strings.NewReader(csvData)
		items, err := ParseBitwardenCSV(reader, true)
		if err != nil {
			t.Fatalf("ParseBitwardenCSV failed: %v", err)
		}
		if len(items) != 5 {
			t.Fatalf("Expected 5 items, got %d", len(items))
		}
		item1 := items[0]
		expectedItem1 := models.ImportedLoginItem{
			SourceName: "Bitwarden", ItemName: "Example Site", Username: "user@example.com", Password: "password123", URL: "https://example.com", Notes: "Some notes here", Folder: "Work", TOTP: "otpauth://totp/Example:user@example.com?secret=JBSWY3DPEHPK3PXP&issuer=Example",
			CustomFields: map[string]string{"type": "login", "reprompt": "0", "favorite": "0", "custom_field_1": "value1", "secret_code": "1234"},
		}
		if !reflect.DeepEqual(item1, expectedItem1) {
			t.Errorf("Item 1 mismatch:\nExpected: %+v\nGot:      %+v", expectedItem1, item1)
		}
		item2 := items[1]
		expectedItem2 := models.ImportedLoginItem{
			SourceName: "Bitwarden", ItemName: "Another Site", Username: "anotheruser", Password: "anotherpass", URL: "https://another.com", Notes: "", Folder: "", TOTP: "",
			CustomFields: map[string]string{"type": "login", "reprompt": "1", "favorite": "1"},
		}
		if !reflect.DeepEqual(item2, expectedItem2) {
			t.Errorf("Item 2 mismatch:\nExpected: %+v\nGot:      %+v", expectedItem2, item2)
		}
		item4 := items[3]
		expectedItem4 := models.ImportedLoginItem{
			SourceName: "Bitwarden", ItemName: "Empty Name", Username: "empty_user", Password: "empty_pass", URL: "", Notes: "", Folder: "Empty Folder", TOTP: "",
			CustomFields: map[string]string{},
		}
		if item4.CustomFields == nil { item4.CustomFields = make(map[string]string) }
		if !reflect.DeepEqual(item4, expectedItem4) {
			t.Errorf("Item 4 mismatch:\nExpected: %+v\nGot:      %+v", expectedItem4, item4)
		}
		item5 := items[4]
		expectedItem5 := models.ImportedLoginItem{
			SourceName: "Bitwarden", ItemName: "Name, with comma", Username: "comma,user", Password: "comma,pass", URL: "https://comma.com", Notes: `Note with "quotes"`, Folder: "Folder, with comma", TOTP: "",
			CustomFields: map[string]string{"type": "login", "reprompt": "0", "favorite": "0"},
		}
		if !reflect.DeepEqual(item5, expectedItem5) {
			t.Errorf("Item 5 (with comma/quotes) mismatch:\nExpected: %+v\nGot:      %+v", expectedItem5, item5)
		}
	})

	t.Run("Parsing without Passwords", func(t *testing.T) {
		reader := strings.NewReader(csvData)
		items, err := ParseBitwardenCSV(reader, false)
		if err != nil {
			t.Fatalf("ParseBitwardenCSV failed: %v", err)
		}
		if len(items) != 5 {
			t.Fatalf("Expected 5 items, got %d", len(items))
		}
		if items[0].Password != "" {
			t.Errorf("Expected empty password for item 1 when importPasswords=false, got '%s'", items[0].Password)
		}
		if items[1].Password != "" {
			t.Errorf("Expected empty password for item 2 when importPasswords=false, got '%s'", items[1].Password)
		}
	})

	t.Run("Handling Invalid or Empty CSV", func(t *testing.T) {
		emptyReader := strings.NewReader("")
		_, err := ParseBitwardenCSV(emptyReader, true)
		if err == nil {
			t.Error("Expected error for empty CSV, got nil")
		} else {
			expectedErr := "CSV 文件为空或没有表头"
			if !strings.Contains(err.Error(), expectedErr) {
				t.Errorf("Expected error message containing '%s', got '%v'", expectedErr, err)
			}
		}
		headerOnlyReader := strings.NewReader("folder,favorite,type,name,notes,fields,reprompt,login_uri,login_username,login_password,login_totp\n")
		items, err := ParseBitwardenCSV(headerOnlyReader, true)
		if err != nil {
			t.Errorf("Expected no error for header-only CSV, got %v", err)
		}
		if len(items) != 0 {
			t.Errorf("Expected 0 items for header-only CSV, got %d", len(items))
		}
		missingColumnReader := strings.NewReader("folder,favorite,type,notes,fields,reprompt,login_uri,login_username,login_password,login_totp\nWork,0,login,note data,,0,https://example.com,user,pass,")
		items, err = ParseBitwardenCSV(missingColumnReader, true)
		if err != nil {
			t.Errorf("Expected no fatal error for missing 'name' column, got %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("Expected 1 item even with missing 'name' column, got %d", len(items))
		}
		if items[0].ItemName != "" {
			t.Errorf("Expected empty ItemName when 'name' column is missing, got '%s'", items[0].ItemName)
		}
		if items[0].Username != "user" {
			t.Errorf("Expected Username 'user', got '%s'", items[0].Username)
		}
	})

	t.Run("Handling Invalid JSON in fields", func(t *testing.T) {
		invalidJsonCsv := `folder,favorite,type,name,notes,fields,reprompt,login_uri,login_username,login_password,login_totp
			,0,login,Invalid JSON Item,"","{""key"": ""value}",0,https://invalid.com,invalid_user,invalid_pass,`
			reader := strings.NewReader(invalidJsonCsv)
		items, err := ParseBitwardenCSV(reader, true)
		if err != nil {
			if !strings.Contains(err.Error(), "JSON") {
				t.Logf("ParseBitwardenCSV returned an unexpected error for invalid JSON: %v", err)
			}
		}
		if len(items) != 1 {
			t.Fatalf("Expected 1 item even with invalid JSON in fields, got %d", len(items))
		}
		expectedRawData := `{"key": "value}`
		if rawData, ok := items[0].CustomFields["_raw_fields_data"]; !ok || rawData != expectedRawData {
			t.Errorf("Expected invalid JSON data '%s' in CustomFields[_raw_fields_data], got: %v", expectedRawData, items[0].CustomFields)
		}
	})

	t.Run("Handling Non-JSON string in fields", func(t *testing.T) {
		nonJsonCsv := `folder,favorite,type,name,notes,fields,reprompt,login_uri,login_username,login_password,login_totp
,0,login,Non-JSON Field Item,"","Just a string",0,https://nonjson.com,nonjson_user,nonjson_pass,`
		reader := strings.NewReader(nonJsonCsv)
		items, err := ParseBitwardenCSV(reader, true)
		if err != nil {
			if !strings.Contains(err.Error(), "JSON") {
				t.Logf("ParseBitwardenCSV returned an unexpected error for non-JSON string: %v", err)
			}
		}
		if len(items) != 1 {
			t.Fatalf("Expected 1 item for non-JSON string in fields, got %d", len(items))
		}
		expectedRawString := "Just a string"
		if rawData, ok := items[0].CustomFields["_raw_fields_data"]; !ok || rawData != expectedRawString {
			t.Errorf("Expected raw string '%s' in CustomFields[_raw_fields_data] when fields is non-JSON, got: %v", expectedRawString, items[0].CustomFields)
		}
	})
}