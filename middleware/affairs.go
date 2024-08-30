package middleware

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net"
	"net/smtp"
	"os"
	"path/filepath"
	"time"

	"github.com/aman1218/portal_bapp/config"
	"github.com/aman1218/portal_bapp/db"
	"github.com/jung-kurt/gofpdf"
	"golang.org/x/exp/rand"
)

func GeneratePDF(data map[string]interface{}, formNumber string) error {

	pdfDir := "pdf"
	if _, err := os.Stat(pdfDir); os.IsNotExist(err) {
		os.Mkdir(pdfDir, 0755)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")

	pdf.SetMargins(15, 10, 10)
	pdf.SetAutoPageBreak(true, 30)

	pdf.SetFont("Arial", "B", 20)

	// Add Page
	pdf.AddPage()

	// Add School Logo
	logoPath := "assets/logo.png" // Replace with your logo's actual path
	if _, err := os.Stat(logoPath); err == nil {
		pdf.ImageOptions(logoPath, 10, 10, 30, 0, false, gofpdf.ImageOptions{ReadDpi: true}, 0, "")
	}

	// Add School Name next to the logo
	pdf.SetXY(50, 15)
	pdf.MultiCell(0, 10, "BOARD OF OPEN SCHOOLING\n& SKILL EDUCATION (B.O.S.S.E.)", "", "L", false)
	pdf.Ln(7) // Slight line break before underline

	// Underline the text
	pdf.CellFormat(0, 10, "", "T", 0, "C", false, 0, "") // "C" centers the underline
	// Line break to start adding data
	pdf.Ln(15)

	pdf.SetFont("Arial", "BU", 18)

	// Center the "Application Details" text
	pdf.SetXY(74, 45)
	pdf.Cell(0, 10, "Application Details")

	// Underline the text

	pdf.Ln(15) // Line break after underline
	// Title
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Student Details:")
	pdf.Ln(12)

	// Add student photo
	if photoFileName, ok := data["studentPhoto"].(string); ok {
		photoPath := filepath.Join("images", photoFileName) // Ensure the path is correctly set
		if _, err := os.Stat(photoPath); err == nil {
			// Define the position and size of the image
			xPos := 170.0     // X position to move the image to the right
			yPos := 50.0      // Y position, adjust as needed
			boxWidth := 35.0  // Width of the image box
			boxHeight := 45.0 // Height of the image box

			pdf.ImageOptions(photoPath, xPos, yPos, boxWidth, boxHeight, false, gofpdf.ImageOptions{ReadDpi: true}, 0, "")
			pdf.SetDrawColor(0, 0, 0) // Set the color to black (R, G, B)
			pdf.Rect(xPos, yPos, boxWidth, boxHeight, "D")
		} else {
			fmt.Println("Student photo not found:", photoPath)
		}
	}

	// Set font for data fields
	pdf.SetFont("Arial", "", 11)

	// Function to add a field as a cell
	addField := func(label string, value string) {
		labelWidth := 50.0
		valueWidth := 100.0

		// Set font to bold for the label
		pdf.SetFont("Arial", "B", 11)
		pdf.CellFormat(labelWidth, 10, label+":", "1", 0, "", false, 0, "")

		// Set font back to regular for the value
		pdf.SetFont("Arial", "", 11)
		pdf.CellFormat(valueWidth, 10, value, "1", 0, "", false, 0, "")

		pdf.Ln(-1) // Moves to the next line after adding a field
	}

	// Function to add array fields as cells
	addArrayField := func(label string, values []interface{}) {
		labelWidth := 50.0
		valueWidth := 150.0

		// Set font to bold for the label
		pdf.SetFont("Arial", "B", 12)
		pdf.CellFormat(labelWidth, 10, label, "1", 0, "", false, 0, "")

		// Set font back to regular for the values
		pdf.SetFont("Arial", "", 12)
		pdf.Ln(-1)
		for _, v := range values {
			pdf.CellFormat(valueWidth, 8, fmt.Sprintf("- %v", v), "1", 0, "", false, 0, "")
			pdf.Ln(8)
		}
		pdf.Ln(2)
	}

	// Student Information
	addField("Form Number", formNumber)
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

	// Education Details
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Education Details:")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 11)
	addField("Board", fmt.Sprintf("%v", data["educationDetails"].(map[string]interface{})["board"]))
	addField("Registration Number", fmt.Sprintf("%v", data["educationDetails"].(map[string]interface{})["registrationNumber"]))
	addField("Roll Number", fmt.Sprintf("%v", data["educationDetails"].(map[string]interface{})["rollNumber"]))
	addField("Year of Passing", fmt.Sprintf("%v", data["educationDetails"].(map[string]interface{})["yearOfPassing"]))

	// Admission Details
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Admission Details:")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 11)
	addField("Admission Mode", fmt.Sprintf("%v", data["admissionMode"]))
	addField("Admission Session", fmt.Sprintf("%v", data["admissionSession"]))
	addField("Course", fmt.Sprintf("%v", data["course"]))
	addField("Category", fmt.Sprintf("%v", data["category"]))

	// Subjects
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Subjects:")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 11)
	if languageSubjects, ok := data["languageSubjects"].([]interface{}); ok {
		addArrayField("Language Subjects", languageSubjects)
	}

	if nonLanguageSubjects, ok := data["nonLanguageSubjects"].([]interface{}); ok {
		addArrayField("Non-Language Subjects", nonLanguageSubjects)
	}

	if vocationalSubjects, ok := data["vocationalSubjects"].([]interface{}); ok {
		addArrayField("Vocational Subjects", vocationalSubjects)
	}

	// Save the PDF file
	err := pdf.OutputFileAndClose(filepath.Join(pdfDir, "application_details.pdf"))
	if err != nil {
		return err
	}

	return nil
}

func HandleEmailtoAdmin(username string, formNumber string) error {
	userDetails, err := db.FetchDetails(username)
	if err != nil {
		fmt.Printf("error fetching name: %v\n", err)

	} else {
		fmt.Println("Name fetched successfully")
	}

	cfg := config.AppConfig
	adminEmail := cfg.Admin.Email

	pdfPath := "pdf/application_details.pdf"

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

	subject := fmt.Sprintf("New Student Admission Details Submitted by %s", userDetails.Name)
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; color: #333;">
			<h2 style="color: #007BFF;">New Student Admission Details</h2>
			<p>Dear Admin,</p>
			<p>A new student has submitted their admission details.</p>
			<p>Form Number: <strong>%s</strong></p>
			<p>Please find attached student details and related images below.</p>
			<p>Best regards,</p>
			<p><strong>panel.org.in</strong></p>
		</body>
		</html>
	`, formNumber)

	err = SendEmail(adminEmail, subject, body, pdfPath, imagePaths)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func HandleEmailtoCoordinator(username string, formNumber string) error {
	userDetails, err := db.FetchDetails(username)
	if err != nil {
		fmt.Printf("error fetching email: %v\n", err)
	} else {
		fmt.Println("Email fetched successfully")
	}

	pdfPath := "pdf/application_details.pdf"

	subject := "New Student Admission Details Submitted"
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; color: #333;">
			<h2 style="color: #007BFF;">New Student Admission Details</h2>
			<p>Dear Coordinator,</p>
			<p>A new student has submitted their admission details.</p>
			<p>Form Number: <strong>%s</strong></p>
			<p>Please find attached student details below.</p>
			<p>Best regards,</p>
			<p><strong>panel.org.in</strong></p>
		</body>
		</html>
	`, formNumber)

	err = SendEmail(userDetails.Email, subject, body, pdfPath, nil)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func HandleEmailtoStudent(data map[string]interface{}, formNumber string) error {
	studentEmail := fmt.Sprintf("%v", data["emailAddress"])

	pdfPath := "pdf/application_details.pdf"

	subject := "BOSSE Admission Details Submitted"
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; color: #333;">
			<h2 style="color: #007BFF;">New Student Admission Details</h2>
			<p>Dear Student,</p>
			<p>Your admission form has been received.</p>
			<p>Form Number: <strong>%s</strong></p>
			<p>Please find attached details below.</p>
			<p>Best regards,</p>
			<p><strong>panel.org.in</strong></p>
		</body>
		</html>
	`, formNumber)

	err := SendEmail(studentEmail, subject, body, pdfPath, nil)
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
	smtpEmail := cfg.Email
	smtpPassword := cfg.Password
	smtpFrom := cfg.From

	// Create the email headers
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", smtpFrom, smtpEmail)
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

	// Connect to the SMTP server (unencrypted connection first)
	conn, err := net.Dial("tcp", smtpHost+":"+smtpPort)
	if err != nil {
		return fmt.Errorf("failed to connect to the server: %v", err)
	}
	defer conn.Close()

	// Create a new SMTP client from the connection
	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}
	defer client.Close()

	// Start TLS encryption (upgrade the connection)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Skip verification for testing, but not recommended for production
		ServerName:         smtpHost,
	}
	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %v", err)
	}

	// Authenticate with the SMTP server
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %v", err)
	}

	// Set the sender and recipient
	if err = client.Mail(smtpEmail); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}
	if err = client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("failed to set recipient: %v", err)
	}

	// Send the email data
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to send data command: %v", err)
	}

	_, err = w.Write(emailBody.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close message: %v", err)
	}

	err = client.Quit()
	if err != nil {
		return fmt.Errorf("failed to quit client: %v", err)
	}

	fmt.Printf("Email successfully sent to %s\n", toEmail)
	return nil
}

func RemoveFiles() {
	// Remove the PDF file
	err := os.Remove("pdf/application_details.pdf")
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

func GenerateUniqueFormNumber() string {
	now := time.Now().Format("200601021504") // Date format: YYYYMMDD-HHMMSS
	rand.Seed(uint64(time.Now().UnixNano()))
	randomNumber := rand.Intn(1000000) // Generate a random number between 0 and 99999
	return fmt.Sprintf("%s-%05d", now, randomNumber)
}
