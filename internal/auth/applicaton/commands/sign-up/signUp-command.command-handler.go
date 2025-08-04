package signup

import (
	"errors"
	"fmt"

	"neuro.app.jordi/internal/auth/domain"
	"neuro.app.jordi/internal/shared/mail"
)

var ErrInvalidCredentials = errors.New("error creating user")

func SignUpCommandHandler(command SignUpCommand, es domain.EncryptionService, userRepo domain.UserRepository, mailService mail.MailProvider) (*domain.User, error) {
	user, err := domain.NewUser(command.Mail, command.Password, command.Name, es)
	if err != nil {
		return nil, err
	}
	if exists := userRepo.Exists(user.Email.Mail); exists {
		return nil, ErrInvalidCredentials
	}
	user.GenerateVerificationCode()

	err = userRepo.Insert(*user)
	if err != nil {
		return nil, err
	}

	// HTML con el código de verificación
	html := fmt.Sprintf(`
<html>
  <body style="font-family: Arial, sans-serif; background:#f0f4f8; padding:20px; margin:0;">
    <div style="max-width:520px; margin:auto; background:white; padding:25px; border-radius:10px; box-shadow:0 2px 8px rgba(0,0,0,0.1);">
      
      <h2 style="color:#007BFF; text-align:center; margin-top:0;">Welcome to <span style="color:#0056b3;">NeuroApp</span>!</h2>
      
      <p style="font-size:15px; color:#333;">Hi %s,</p>
      <p style="font-size:15px; color:#333; line-height:1.5;">
        Thank you for signing up. Use the verification code below to activate your account:
      </p>
      
      <div style="background:#007BFF; color:white; font-size:20px; font-weight:bold; padding:12px; text-align:center; border-radius:6px; letter-spacing:2px; margin:20px 0;">
        %s
      </div>
      
      <p style="font-size:14px; color:#555; line-height:1.5;">
        If you didn’t create this account, you can safely ignore this email.
      </p>
      
      <hr style="border:none; border-top:1px solid #e6e6e6; margin:25px 0;">
      <p style="font-size:12px; color:#999; text-align:center;">
        © 2025 NeuroApp. All rights reserved.
      </p>
    </div>
  </body>
</html>
`, user.Name, user.VerificationCode)

	err = mailService.SendMailHTML(user.Email.Mail, "Welcome to NeuroApp", html)
	return user, err
}
