package signup

import (
	"context"
	"errors"
	"fmt"
	"time"

	"neuro.app.jordi/internal/auth/domain"
	"neuro.app.jordi/internal/shared/mail"
)

var ErrInvalidCredentials = errors.New("error creating user")

func SignUpCommandHandler(ctx context.Context, command SignUpCommand, userRepo domain.UserRepository, mailService mail.MailProvider) (*domain.User, string, error) {
	user, err := domain.NewUser(command.Name, command.Mail)
	if err != nil {
		return nil, "", err
	}
	if exists := userRepo.Exists(ctx, user.Email.Mail); exists {
		return nil, "", ErrInvalidCredentials
	}
	err = userRepo.Insert(ctx, *user)
	if err != nil {
		return nil, "", err
	}

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <style>
    body { font-family: Arial, sans-serif; background-color:#f9f9f9; padding:20px; }
    .container { max-width:600px; margin:auto; background:#ffffff; border-radius:8px; padding:30px; box-shadow:0 2px 6px rgba(0,0,0,0.1); }
    h1 { color:#2c3e50; }
    p { color:#555555; line-height:1.5; }
    .footer { margin-top:30px; font-size:12px; color:#888888; text-align:center; }
  </style>
</head>
<body>
  <div class="container">
    <h1>Welcome to <span style="color:#3B82F6;">NeuroApp</span>!</h1>
    <p>Hi %s,</p>
    <p>We’re delighted to have you on board. NeuroApp was designed to help healthcare professionals and patients work together with ease and confidence.</p>
    <p>You can now start exploring the platform and discover how it can support your work and improve your daily routine.</p>
    <p>If you have any questions or need assistance, don’t hesitate to reach out to our support team.</p>
    <p>We wish you a great start with NeuroApp!</p>
    <div class="footer">
      © %d NeuroApp. All rights reserved.
    </div>
  </div>
</body>
</html>
`, user.Name.Name, time.Now().Year())

	err = mailService.SendEmail(
		user.Email.Mail,
		"Welcome to NeuroApp",
		body, // HTML content
		"",
		nil,
	)
	return nil, "", err
}
