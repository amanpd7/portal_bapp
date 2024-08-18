package middleware

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"os"
	"path/filepath"

	"github.com/aman1218/portal_bapp/config"
	"github.com/aman1218/portal_bapp/db"
	"github.com/jung-kurt/gofpdf"
)

func GeneratePDF(data map[string]interface{}) error {

	pdfDir := "pdf"
	if _, err := os.Stat(pdfDir); os.IsNotExist(err) {
		os.Mkdir(pdfDir, 0755)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")

	pdf.SetFont("Arial", "B", 16)

	pdf.AddPage()

	pdf.Cell(0, 10, "Student Admission Details")
	pdf.Ln(12)

	if photoFileName, ok := data["studentPhoto"].(string); ok {
		photoPath := filepath.Join("images", photoFileName)
		if _, err := os.Stat(photoPath); err == nil { // Check if file exists
			pdf.ImageOptions(photoPath, 150, 10, 40, 0, false, gofpdf.ImageOptions{ReadDpi: true}, 0, "")
		} else {
			fmt.Println("Student photo not found:", photoPath)
		}
	}

	pdf.SetFont("Arial", "", 12)

	addField := func(label string, value string) {
		labelWidth := 50.0
		valueWidth := 100.0

		pdf.Cell(labelWidth, 10, label+":")
		pdf.Cell(valueWidth, 10, value)
		pdf.Ln(10)
	}

	addArrayField := func(label string, values []interface{}) {
		labelWidth := 50.0

		pdf.Cell(labelWidth, 10, label+":")
		pdf.Ln(10)
		for _, v := range values {
			pdf.Cell(labelWidth+10, 8, fmt.Sprintf("- %v", v))
			pdf.Ln(8)
		}
	}

	addField("Student Name", fmt.Sprintf("%v", data["studentName"]))
	addField("Father's Name", fmt.Sprintf("%v", data["fatherName"]))
	addField("Mother's Name", fmt.Sprintf("%v", data["motherName"]))
	addField("Gender", fmt.Sprintf("%v", data["gender"]))
	addField("Date of Birth", fmt.Sprintf("%s/%s/%s", data["dob"].(map[string]interface{})["day"], data["dob"].(map[string]interface{})["month"], data["dob"].(map[string]interface{})["year"]))
	addField("Contact Number", fmt.Sprintf("%v", data["contactNumber"]))
	addField("Email Address", fmt.Sprintf("%v", data["emailAddress"]))
	addField("Permanent Address", fmt.Sprintf("%v", data["permanentAddress"]))
	addField("City", fmt.Sprintf("%v", data["city"]))
	addField("State", fmt.Sprintf("%v", data["state"]))
	addField("Country", fmt.Sprintf("%v", data["country"]))
	addField("Pin Code", fmt.Sprintf("%v", data["pinCode"]))
	addField("Aadhar Number", fmt.Sprintf("%v", data["aadharNumber"]))

	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Education Details")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	addField("Board", fmt.Sprintf("%v", data["educationDetails"].(map[string]interface{})["board"]))
	addField("Registration Number", fmt.Sprintf("%v", data["educationDetails"].(map[string]interface{})["registrationNumber"]))
	addField("Roll Number", fmt.Sprintf("%v", data["educationDetails"].(map[string]interface{})["rollNumber"]))
	addField("Year of Passing", fmt.Sprintf("%v", data["educationDetails"].(map[string]interface{})["yearOfPassing"]))

	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Admission Details")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	addField("Admission Mode", fmt.Sprintf("%v", data["admissionMode"]))
	addField("Admission Session", fmt.Sprintf("%v", data["admissionSession"]))
	addField("Course", fmt.Sprintf("%v", data["course"]))
	addField("Category", fmt.Sprintf("%v", data["category"]))

	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Subjects")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	if languageSubjects, ok := data["languageSubjects"].([]interface{}); ok {
		addArrayField("Language Subjects", languageSubjects)
	}

	if nonLanguageSubjects, ok := data["nonLanguageSubjects"].([]interface{}); ok {
		addArrayField("Non-Language Subjects", nonLanguageSubjects)
	}

	if vocationalSubjects, ok := data["vocationalSubjects"].([]interface{}); ok {
		addArrayField("Vocational Subjects", vocationalSubjects)
	}

	pdf.Ln(5)

	err := pdf.OutputFileAndClose("pdf/student_details.pdf")
	if err != nil {
		return err
	}

	return nil
}

func HandleEmailtoAdmin(username string) error {
	userbyName, err := db.FetchName(username)
	if err != nil {
		fmt.Printf("error fetching name: %v\n", err)
	} else {
		fmt.Println("Name fetched successfully")
	}

	cfg := config.AppConfig
	adminEmail := cfg.Admin.Email

	pdfPath := "pdf/student_details.pdf"

	uploadsDir := "images"
	files, err := os.ReadDir(uploadsDir)
	if err != nil {
		fmt.Printf("error reading uploads directory: %v", err)
	}

	var imagePaths []string
	for _, file := range files {
		if !file.IsDir() {
			imagePaths = append(imagePaths, filepath.Join(uploadsDir, file.Name()))
		}
	}

	imagePaths = append(imagePaths, pdfPath)

	subject := fmt.Sprintf("New Student Admission Details Submitted by %s", userbyName.Name)
	body := `
		<html>
		<body style="font-family: Arial, sans-serif; color: #333;">
			<h2 style="color: #007BFF;">New Student Admission Details</h2>
			<p>Dear Admin,</p>
			<p>A new student has submitted their admission details.</p>
			<p>Please find attached the student details and related images.</p>
			<p>Best regards,</p>
			<p><strong>panel.org.in</strong></p>
		</body>
		</html>
	`

	err = SendEmail(adminEmail, subject, body, pdfPath, imagePaths)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func SendEmail(toEmail string, subject string, body string, pdfPath string, imagePaths []string) error {
	cfg := config.AppConfig.Smtp
	smtpHost := cfg.Host
	smtpPort := cfg.Port
	smtpUser := cfg.User
	smtpPassword := cfg.Password

	// Create the email headers
	headers := make(map[string]string)
	headers["From"] = smtpUser
	headers["To"] = toEmail
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = `multipart/mixed; boundary="boundary"`

	// Buffer to write the email content
	var emailBody bytes.Buffer

	// Write the headers to the buffer
	for k, v := range headers {
		emailBody.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	emailBody.WriteString("\r\n--boundary\r\n")

	// Write the email body part
	emailBody.WriteString(`Content-Type: text/html; charset="utf-8"` + "\r\n")
	emailBody.WriteString("\r\n" + body + "\r\n")
	emailBody.WriteString("--boundary\r\n")

	// Function to attach a file
	attachFile := func(filePath string) error {
		fileName := filepath.Base(filePath)

		// Open the file
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		// Read file content
		fileData, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}

		// Write attachment header
		emailBody.WriteString("Content-Type: application/octet-stream\r\n")
		emailBody.WriteString("Content-Transfer-Encoding: base64\r\n")
		emailBody.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", fileName))
		emailBody.WriteString("\r\n")

		// Write attachment content
		encoded := base64.StdEncoding.EncodeToString(fileData)
		emailBody.WriteString(encoded)
		emailBody.WriteString("\r\n--boundary\r\n")

		return nil
	}

	// Attach the PDF
	err := attachFile(pdfPath)
	if err != nil {
		return err
	}

	// Attach the images
	for _, imagePath := range imagePaths {
		err := attachFile(imagePath)
		if err != nil {
			return err
		}
	}

	// Convert buffer to bytes and send the email
	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpHost)
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, []string{toEmail}, emailBody.Bytes())

	if err != nil {
		return err
	}

	return nil
}

func RemoveFiles() {
	// Remove the PDF file
	err := os.Remove("pdf/student_details.pdf")
	if err != nil {
		fmt.Printf("Error removing PDF file: %v\n", err)
	}

	// Remove all image files
	uploadsDir := "images"
	files, err := os.ReadDir(uploadsDir)
	if err != nil {
		fmt.Printf("Error reading uploads directory: %v\n", err)
	}

	for _, file := range files {
		if !file.IsDir() { // Ensure it's a file, not a directory
			err := os.Remove(filepath.Join(uploadsDir, file.Name()))
			if err != nil {
				fmt.Printf("Error removing image file: %v\n", err)
			}
		}
	}
}
