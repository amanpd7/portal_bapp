package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aman1218/portal_bapp/db"
	"github.com/aman1218/portal_bapp/middleware"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := db.Authenticate(req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := middleware.GenerateJWT(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username   string `json:"username"`
		Password   string `json:"password"`
		Email      string `json:"email"`
		Name       string `json:"name"`
		SecretPass string `json:"secretpass"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.SecretPass == "portal_bapp" {
		_, err := db.Register(req.Username, req.Password, req.Email, req.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Invalid secret", http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func FormHandler(w http.ResponseWriter, r *http.Request) {

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := middleware.ValidateToken(tokenString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Use the correct field based on your claims structure
	username := claims.Username // Assuming your JWT has a "username" claim
	if username == "" {
		username = claims.Subject // Fall back to "sub" if "username" is not set
	}
	if username == "" {
		http.Error(w, "Unable to extract username", http.StatusUnauthorized)
		return
	}

	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	formData := make(map[string]interface{})

	for key, values := range r.MultipartForm.Value {
		if len(values) > 0 {
			switch key {
			case "dob", "educationDetails", "languageSubjects", "nonLanguageSubjects", "vocationalSubjects":
				var jsonValue interface{}
				err := json.Unmarshal([]byte(values[0]), &jsonValue)
				if err != nil {
					fmt.Printf("Error parsing JSON field %s: %v\n", key, err)
					formData[key] = values[0]
				} else {
					formData[key] = jsonValue
				}
			default:
				formData[key] = values[0]
			}
		}
	}

	uploadDir := "images"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}

	files := []string{"document_AADHAR", "document_X", "document_XII", "studentPhoto"}

	for _, fileKey := range files {
		file, handler, err := r.FormFile(fileKey)
		if err != nil {
			fmt.Println("Error retrieving the file:", fileKey, err)
			continue
		}
		defer file.Close()

		dst, err := os.Create(filepath.Join(uploadDir, handler.Filename))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		formData[fileKey] = handler.Filename

		fmt.Printf("File %s uploaded successfully\n", handler.Filename)
	}

	jsonData, err := json.MarshalIndent(formData, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data map[string]interface{}
	err = json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return
	}

	go func(data map[string]interface{}, username string) {
		err := middleware.GeneratePDF(data)
		if err != nil {
			fmt.Printf("Error generating PDF: %v\n", err)
		} else {
			fmt.Println("PDF generated successfully!")
		}

		err = middleware.HandleEmailtoStudent(data)
		if err != nil {
			fmt.Printf("Error sending email to student: %v\n", err)
		} else {
			fmt.Println("Email to student sent successfully!")
		}

		err = middleware.HandleEmailtoAdmin(username)
		if err != nil {
			fmt.Printf("Error sending email to admin: %v\n", err)
		} else {
			fmt.Println("Email to admin sent successfully!")
		}

		middleware.RemoveFiles()
		
	}(data, username)

}