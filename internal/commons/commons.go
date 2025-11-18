package commons

import (
	"log"
	"net/smtp"
	"strings"
)

const (

	// Retroactive questions
	DAPTBR   int = 200 // 65
	DAENUS   int = 50  // 15
	DEEPPTBR int = 50  // 7
	DEEPENUS int = 50  // 2
	DEEPESES int = 50  // 2
	RADDLE   int = 50  // 2

	// Date format
	YYYYMMDD       = "2006-01-02"
	YYYYMMDDhhmmss = "2006-01-02-15:04:05"

	TO       = ""
	FROM     = ""
	PASSWORD = ""
	SMTPHOST = "smtp.gmail.com"
	SMTPPORT = "587"
)

type Error struct {
	Err    error  `json:"-"`
	Msg    string `json:"message"`
	Status int    `json:"-"`
}

type Done struct {
	Msg    string `json:"message"`
	Status int    `json:"-"`
}

func SendEmail(log *log.Logger) {
	// Receiver email address.
	to := []string{TO}

	// Authentication.
	auth := smtp.PlainAuth("", FROM, PASSWORD, SMTPHOST)

	// Message.
	message := []byte("To: anom@gmail.com\r\n" +

		"Subject: Error on API MSc Project\r\n" +

		"API MSC Email Error, please take a look!\r\n")

	// Sending email.
	err := smtp.SendMail(SMTPHOST+":"+SMTPPORT, auth, FROM, to, message)
	if err != nil {
		log.Println("error on sending email, error: ", err)
	}
	log.Println("Email sent successfully!")
}

func GetStringInBetweenTwoString(value string, startString string, endString string) (result string) {
	newFirstPos := strings.Index(value, startString)
	if newFirstPos == -1 {
		return result
	}
	newValue := value[newFirstPos+len(startString):]
	newLastPos := strings.Index(newValue, endString)
	if newLastPos == -1 {
		return result
	}
	result = newValue[:newLastPos]
	return result
}
