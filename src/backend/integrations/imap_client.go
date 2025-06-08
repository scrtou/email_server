package integrations

import (
	"context"
	"email_server/database"
	"email_server/models"
	"email_server/utils"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type xoauth2Client struct {
	user  string
	token string
}

func (a *xoauth2Client) Start() (string, []byte, error) {
	// The client-first mechanism for XOAUTH2 requires sending the initial
	// response formatted as "user=<user>\x01auth=Bearer <token>\x01\x01"
	initialResponse := []byte(fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", a.user, a.token))
	return "XOAUTH2", initialResponse, nil
}

func (a *xoauth2Client) Next(challenge []byte) ([]byte, error) {
	// XOAUTH2 is a client-first mechanism, so we don't expect a challenge.
	// If we receive one, it's likely an error from the server.
	// The spec says the client should respond with a single CRLF.
	return nil, fmt.Errorf("unexpected server challenge: %s", string(challenge))
}

// connectAndLogin handles the connection and authentication logic.
func connectAndLogin(emailAccount models.EmailAccount, token *models.UserOAuthToken) (*imapclient.Client, error) {
	var imapServer, imapPort = emailAccount.IMAPServer, emailAccount.IMAPPort

	// If it's an OAuth2 connection, we need to fetch the provider's IMAP details from the DB.
	if token != nil {
		var provider models.OAuthProvider
		if err := database.DB.First(&provider, token.ProviderID).Error; err != nil {
			return nil, fmt.Errorf("failed to find oauth provider with id %d for imap details: %w", token.ProviderID, err)
		}
		imapServer = provider.IMAPServer
		imapPort = provider.IMAPPort
	}

	imapServerAddr := fmt.Sprintf("%s:%d", imapServer, imapPort)
	authMethod := "Password"
	if token != nil {
		authMethod = "XOAUTH2"
	}
	log.Printf("Connecting to IMAP server: %s for user %s (Auth: %s)", imapServerAddr, emailAccount.EmailAddress, authMethod)

	c, err := imapclient.DialTLS(imapServerAddr, &imapclient.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to IMAP server: %w", err)
	}

	if token != nil {
		// OAuth2 authentication
		decryptedAccessToken, err := utils.Decrypt(token.AccessTokenEncrypted)
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("could not decrypt access token: %w", err)
		}

		decryptedRefreshToken, err := utils.Decrypt(token.RefreshTokenEncrypted)
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("could not decrypt refresh token: %w", err)
		}

		oauth2Token := &oauth2.Token{
			AccessToken:  string(decryptedAccessToken),
			RefreshToken: string(decryptedRefreshToken),
			Expiry:       token.Expiry,
			TokenType:    token.TokenType,
		}

		// Check if the token is expired and refresh if necessary
		if !oauth2Token.Valid() {
			log.Printf("OAuth2 token for %s is expired, refreshing...", emailAccount.EmailAddress)

			var provider models.OAuthProvider
			if err := database.DB.First(&provider, token.ProviderID).Error; err != nil {
				c.Close()
				return nil, fmt.Errorf("failed to find oauth provider with id %d: %w", token.ProviderID, err)
			}

			decryptedSecret, err := utils.Decrypt(provider.ClientSecretEncrypted)
			if err != nil {
				c.Close()
				return nil, fmt.Errorf("could not decrypt provider client secret: %w", err)
			}

			conf := &oauth2.Config{
				ClientID:     provider.ClientID,
				ClientSecret: string(decryptedSecret),
				Scopes:       strings.Split(provider.Scopes, ","), // <-- 多了这一行
				Endpoint: oauth2.Endpoint{
					AuthURL:  provider.AuthURL,
					TokenURL: provider.TokenURL,
				},
			}

			tokenSource := conf.TokenSource(context.Background(), oauth2Token)
			newToken, err := tokenSource.Token()
			if err != nil {
				c.Close()
				return nil, fmt.Errorf("failed to refresh token: %w", err)
			}

			// Update the local token variable and the database
			oauth2Token = newToken
			token.AccessTokenEncrypted, err = utils.Encrypt([]byte(newToken.AccessToken))
			if err != nil {
				c.Close()
				return nil, fmt.Errorf("failed to encrypt new access token: %w", err)
			}
			// Only update refresh token if a new one was provided
			if newToken.RefreshToken != "" {
				token.RefreshTokenEncrypted, err = utils.Encrypt([]byte(newToken.RefreshToken))
				if err != nil {
					c.Close()
					return nil, fmt.Errorf("failed to encrypt new refresh token: %w", err)
				}
			}
			token.Expiry = newToken.Expiry

			if err := database.DB.Save(token).Error; err != nil {
				c.Close()
				return nil, fmt.Errorf("failed to save updated token to database: %w", err)
			}
			log.Printf("Successfully refreshed and saved new token for %s", emailAccount.EmailAddress)
		}

		log.Printf("Attempting XOAUTH2 login for %s", emailAccount.EmailAddress)
		log.Printf("--- FULL ACCESS TOKEN FOR DEBUG ---\n%s\n---------------------------------", oauth2Token.AccessToken)
		auth := &xoauth2Client{
			user:  emailAccount.EmailAddress,
			token: oauth2Token.AccessToken,
		}
		if err := c.Authenticate(auth); err != nil {
			c.Close()
			return nil, fmt.Errorf("XOAUTH2 login failed: %w", err)
		}
		log.Printf("Successfully logged in with XOAUTH2 for %s", emailAccount.EmailAddress)
	} else {
		// Password (PLAIN) authentication
		password, err := utils.DecryptPassword(emailAccount.PasswordEncrypted)
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("could not decrypt password: %w", err)
		}
		if password == "" {
			c.Close()
			return nil, fmt.Errorf("password for email account %s is not set", emailAccount.EmailAddress)
		}

		if err := c.Login(emailAccount.EmailAddress, password).Wait(); err != nil {
			c.Close()
			return nil, fmt.Errorf("IMAP login failed: %w", err)
		}
		log.Printf("Successfully logged in with password for %s", emailAccount.EmailAddress)
	}

	return c, nil
}

// FetchEmails connects to an IMAP server and fetches emails with pagination.
func FetchEmails(emailAccount models.EmailAccount, page, pageSize int) ([]models.Email, int, error) {
	var token models.UserOAuthToken
	err := database.DB.Where("email_account_id = ?", emailAccount.ID).First(&token).Error

	var c *imapclient.Client
	if err == nil {
		// Token found, use OAuth2
		c, err = connectAndLogin(emailAccount, &token)
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// No token found, use password
		log.Printf("No OAuth token found for %s, falling back to password authentication.", emailAccount.EmailAddress)
		c, err = connectAndLogin(emailAccount, nil)
	} else {
		// Database error
		return nil, 0, fmt.Errorf("failed to query for oauth token: %w", err)
	}

	if err != nil {
		log.Printf("Failed to connect or login for %s: %v", emailAccount.EmailAddress, err)
		return nil, 0, err
	}
	defer c.Close()

	// Select INBOX
	mailbox, err := c.Select("INBOX", nil).Wait()
	if err != nil {
		log.Printf("Failed to select INBOX for %s: %v", emailAccount.EmailAddress, err)
		return nil, 0, fmt.Errorf("failed to select INBOX: %w", err)
	}
	totalMessages := int(mailbox.NumMessages)
	log.Printf("INBOX selected. Total messages: %d", totalMessages)

	if totalMessages == 0 {
		return []models.Email{}, 0, nil
	}

	// Calculate message sequence numbers for the requested page
	start := totalMessages - (page * pageSize) + 1
	end := totalMessages - ((page - 1) * pageSize)

	if start < 1 {
		start = 1
	}
	if end < 1 || start > end {
		log.Printf("No messages to fetch for page %d, pageSize %d", page, pageSize)
		return []models.Email{}, totalMessages, nil
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddRange(uint32(start), uint32(end))
	log.Printf("Fetching messages in range: %d-%d", start, end)

	fetchOptions := &imap.FetchOptions{
		Envelope: true,
		Flags:    true,
		// Fetching body sections can be added here if needed
	}

	// Fetch messages using Collect, which waits for the command to finish.
	messages, err := c.Fetch(seqSet, fetchOptions).Collect()
	if err != nil {
		log.Printf("Failed to fetch emails for %s: %v", emailAccount.EmailAddress, err)
		return nil, 0, fmt.Errorf("IMAP fetch command failed: %w", err)
	}

	var emails []models.Email
	// Reverse the order of messages to have the newest first
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if msg == nil {
			log.Println("Received a nil message from fetch command")
			continue
		}

		from := make([]models.EmailAddress, len(msg.Envelope.From))
		for i, addr := range msg.Envelope.From {
			from[i] = models.EmailAddress{Name: addr.Mailbox, Address: addr.Host}
		}

		to := make([]models.EmailAddress, len(msg.Envelope.To))
		for i, addr := range msg.Envelope.To {
			to[i] = models.EmailAddress{Name: addr.Mailbox, Address: addr.Host}
		}

		date := time.Time{}
		if !msg.Envelope.Date.IsZero() {
			date = msg.Envelope.Date
		}

		email := models.Email{
			MessageID: msg.Envelope.MessageID,
			Subject:   msg.Envelope.Subject,
			From:      from,
			To:        to,
			Date:      date,
		}
		emails = append(emails, email)
	}

	log.Printf("Successfully fetched %d emails for %s", len(emails), emailAccount.EmailAddress)
	return emails, totalMessages, nil
}
