package main

import (
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"io"
	"io/ioutil"
	"log"
)

func main() {
	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer func(c *client.Client) {
		err := c.Logout()
		if err != nil {

		}
	}(c)

	// Login
	if err := c.Login("theotruvelott@gmail.com", "quzkwxfmzztgfmdw"); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	_, err = c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	ids, err := c.Search(criteria)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("IDs found:", ids)

	if len(ids) > 0 {
		seqset := new(imap.SeqSet)
		seqset.AddNum(ids...)
		var section imap.BodySectionName
		items := []imap.FetchItem{section.FetchItem()}

		messages := make(chan *imap.Message, 1)
		done := make(chan error, 1)
		go func() {
			done <- c.Fetch(seqset, items, messages)
		}()

		log.Println("Unseen messages:")
		counter := 0
		for msg := range messages {
			counter++
			log.Println("Message numéro ", counter)
			r := msg.GetBody(&section)
			if r == nil {
				log.Fatal("Server didn't returned message body")
			}

			// Create a new mail reader
			mr, err := mail.CreateReader(r)
			if err != nil {
				log.Fatal(err)
			}

			// Print some info about the message
			header := mr.Header
			if date, err := header.Date(); err == nil {
				log.Println("Date:", date)
			}
			if from, err := header.AddressList("From"); err == nil {
				log.Println("From:", from)
			}
			if to, err := header.AddressList("To"); err == nil {
				log.Println("To:", to)
			}
			if subject, err := header.Subject(); err == nil {
				log.Println("Subject:", subject)
			}

			// Process each message's part (in this case plain text)
			for i := 0; i < 1; {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				} else if err != nil {
					log.Fatal(err)
				}

				switch o := p.Header.(type) {

				case *mail.InlineHeader:
					log.Println(o)
					b, _ := ioutil.ReadAll(p.Body)
					if b != nil {
						log.Println("Body: \n ", string(b))
						i++
						break

					}

				}

			}
			// same but for attachments
			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				} else if err != nil {
					log.Fatal(err)
				}
				switch h := p.Header.(type) {
				case *mail.AttachmentHeader:
					// This is an attachment
					filename, _ := h.Filename()

					log.Println("Got attachment: ", filename)
					// Save attachment to ./attachments

					file, _ := ioutil.ReadAll(p.Body)
					if err := ioutil.WriteFile("./attachments/"+filename, file, 0666); err != nil {
						log.Fatal(err)
					}
				}

			}
			//move mail to another folder

			if err := c.Store(seqset, "X-GM-LABELS", []interface{}{"\\Trash"}, nil); err != nil {
				log.Fatal(err)
			}
			log.Println("")
		}

		log.Println("Done!")

	}
}
