package gqlt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInput_LoadQuery(t *testing.T) {
	input := NewInput()

	tests := []struct {
		name        string
		query       string
		queryFile   string
		wantErr     bool
		setupFile   func() string // returns filename to cleanup
		cleanupFile func(string)
	}{
		{
			name:      "inline query",
			query:     "{ users { id name } }",
			queryFile: "",
			wantErr:   false,
		},
		{
			name:      "query from file",
			query:     "",
			queryFile: "",
			wantErr:   false,
			setupFile: func() string {
				f, _ := os.CreateTemp("", "test-query-*.graphql")
				f.WriteString("{ users { id name } }")
				f.Close()
				return f.Name()
			},
			cleanupFile: func(filename string) {
				os.Remove(filename)
			},
		},
		{
			name:      "both query and file specified",
			query:     "{ users { id } }",
			queryFile: "",
			wantErr:   false, // LoadQuery doesn't validate this, it just returns the query
			setupFile: func() string {
				f, _ := os.CreateTemp("", "test-query-*.graphql")
				f.WriteString("{ users { id name } }")
				f.Close()
				return f.Name()
			},
			cleanupFile: func(filename string) {
				os.Remove(filename)
			},
		},
		{
			name:      "neither query nor file specified",
			query:     "",
			queryFile: "",
			wantErr:   true,
		},
		{
			name:      "file not found",
			query:     "",
			queryFile: "/nonexistent/file.graphql",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryFile := tt.queryFile
			if tt.setupFile != nil {
				queryFile = tt.setupFile()
				if tt.cleanupFile != nil {
					defer tt.cleanupFile(queryFile)
				}
			}

			got, err := input.LoadQuery(tt.query, queryFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Errorf("LoadQuery() returned empty string")
			}
		})
	}
}

func TestInput_LoadVariables(t *testing.T) {
	input := NewInput()

	tests := []struct {
		name      string
		vars      string
		varsFile  string
		wantErr   bool
		setupFile func() string
		cleanup   func(string)
	}{
		{
			name:     "inline variables",
			vars:     `{"id": 1, "name": "test"}`,
			varsFile: "",
			wantErr:  false,
		},
		{
			name:     "variables from file",
			vars:     "",
			varsFile: "",
			wantErr:  false,
			setupFile: func() string {
				f, _ := os.CreateTemp("", "test-vars-*.json")
				f.WriteString(`{"id": 1, "name": "test"}`)
				f.Close()
				return f.Name()
			},
			cleanup: func(filename string) {
				os.Remove(filename)
			},
		},
		{
			name:     "invalid JSON",
			vars:     `{"id": 1, "name": "test"`,
			varsFile: "",
			wantErr:  true,
		},
		{
			name:     "file not found",
			vars:     "",
			varsFile: "/nonexistent/vars.json",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			varsFile := tt.varsFile
			if tt.setupFile != nil {
				varsFile = tt.setupFile()
				if tt.cleanup != nil {
					defer tt.cleanup(varsFile)
				}
			}

			got, err := input.LoadVariables(tt.vars, varsFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("LoadVariables() returned nil map")
			}
		})
	}
}

func TestInput_LoadHeaders(t *testing.T) {
	input := NewInput()

	tests := []struct {
		name    string
		headers []string
		want    map[string]string
	}{
		{
			name:    "single header",
			headers: []string{"Authorization: Bearer token123"},
			want:    map[string]string{"Authorization": "Bearer token123"},
		},
		{
			name:    "multiple headers",
			headers: []string{"Authorization: Bearer token123", "Content-Type: application/json"},
			want:    map[string]string{"Authorization": "Bearer token123", "Content-Type": "application/json"},
		},
		{
			name:    "empty headers",
			headers: []string{},
			want:    map[string]string{},
		},
		{
			name:    "malformed header",
			headers: []string{"InvalidHeader"},
			want:    map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := input.LoadHeaders(tt.headers)
			if len(got) != len(tt.want) {
				t.Errorf("LoadHeaders() got %d headers, want %d", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("LoadHeaders() header %s = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}

func TestInput_ParseFiles(t *testing.T) {
	input := NewInput()

	// Create temporary files for testing
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "test1.txt")
	file2 := filepath.Join(tempDir, "test2.txt")

	os.WriteFile(file1, []byte("content1"), 0644)
	os.WriteFile(file2, []byte("content2"), 0644)

	tests := []struct {
		name    string
		files   []string
		wantErr bool
	}{
		{
			name:    "valid files",
			files:   []string{"file1=" + file1, "file2=" + file2},
			wantErr: false,
		},
		{
			name:    "nonexistent file",
			files:   []string{"file1=/nonexistent/file.txt"},
			wantErr: true,
		},
		{
			name:    "malformed file spec",
			files:   []string{"invalid"},
			wantErr: true,
		},
		{
			name:    "empty files",
			files:   []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := input.ParseFiles(tt.files)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("ParseFiles() returned nil map")
			}
		})
	}
}

func TestInput_ParseFilesFromList(t *testing.T) {
	input := NewInput()

	// Create temporary files for testing
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "test1.txt")
	file2 := filepath.Join(tempDir, "test2.txt")

	os.WriteFile(file1, []byte("content1"), 0644)
	os.WriteFile(file2, []byte("content2"), 0644)

	// Create files list
	filesList := filepath.Join(tempDir, "files.txt")
	os.WriteFile(filesList, []byte("file1="+file1+"\nfile2="+file2+"\n"), 0644)

	tests := []struct {
		name      string
		filesList string
		wantErr   bool
	}{
		{
			name:      "valid files list",
			filesList: filesList,
			wantErr:   false,
		},
		{
			name:      "nonexistent files list",
			filesList: "/nonexistent/files.txt",
			wantErr:   true,
		},
		{
			name:      "empty files list",
			filesList: "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := input.ParseFilesFromList(tt.filesList)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFilesFromList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("ParseFilesFromList() returned nil slice")
			}
		})
	}
}
