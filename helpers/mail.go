package helpers

import (
	"net/smtp"
	"os"

	"github.com/jordan-wright/email"
)

func SendForgotPasswordEmail(token, mail string) error {
	url := (os.Getenv("BASE_URI") + "api/v1/confirm-password-reset?token=" + token + "&mail=" + mail)

	e := email.NewEmail()
	e.From = "Watchlistfy <" + os.Getenv("FROM_MAIL") + ">"
	e.To = []string{mail}
	e.Subject = "Forgot Password"
	e.HTML = []byte(
		`<!doctype html>
		<html lang="en-US">

		<head>
			<meta content="text/html; charset=utf-8" http-equiv="Content-Type" />
			<title>Reset Password</title>
			<meta name="description" content="Reset Password.">
			<style type="text/css">
				a:hover {text-decoration: underline !important;}
			</style>
		</head>

		<body marginheight="0" topmargin="0" marginwidth="0" style="margin: 0px; background-color: #f2f3f8;" leftmargin="0">
			<table cellspacing="0" border="0" cellpadding="0" width="100%" bgcolor="#f2f3f8"
				style="@import url(https://fonts.googleapis.com/css?family=Rubik:300,400,500,700|Open+Sans:300,400,600,700); font-family: 'Open Sans', sans-serif;">
				<tr>
					<td>
						<table style="background-color: #f2f3f8; max-width:670px;  margin:0 auto;" width="100%" border="0"
							align="center" cellpadding="0" cellspacing="0">
							<tr>
								<td style="height:80px;">&nbsp;</td>
							</tr>
							<tr>
								<td style="text-align:center;">
									<img width="125" src="https://user-images.githubusercontent.com/25686023/269638549-fc57304b-bc39-4e66-ae26-7cfc59408a21.png" title="logo" alt="logo">
								</td>
							</tr>
							<tr>
								<td style="height:20px;">&nbsp;</td>
							</tr>
							<tr>
								<td>
									<table width="95%" border="0" align="center" cellpadding="0" cellspacing="0"
										style="max-width:670px;background:#fff; border-radius:3px; text-align:center;-webkit-box-shadow:0 6px 18px 0 rgba(0,0,0,.06);-moz-box-shadow:0 6px 18px 0 rgba(0,0,0,.06);box-shadow:0 6px 18px 0 rgba(0,0,0,.06);">
										<tr>
											<td style="height:40px;">&nbsp;</td>
										</tr>
										<tr>
											<td style="padding:0 35px;">
												<h1 style="color:#1e1e2d; font-weight:500; margin:0;font-size:32px;font-family:'Rubik',sans-serif;">You have
													requested to reset your password</h1>
												<span
													style="display:inline-block; vertical-align:middle; margin:29px 0 26px; border-bottom:1px solid #cecece; width:100px;"></span>
												<p style="color:#455056; font-size:15px;line-height:24px; margin:0;">
													We cannot simply send you your old password. A unique link to reset your
													password has been generated for you. To reset your password, click the
													following link and follow the instructions.
												</p>
												<a href="` + url + `"
													style="background:#20e277;text-decoration:none !important; font-weight:500; margin-top:35px; color:#fff;text-transform:uppercase; font-size:14px;padding:10px 24px;display:inline-block;border-radius:50px;">Reset
													Password</a>
											</td>
										</tr>
										<tr>
											<td style="height:40px;">&nbsp;</td>
										</tr>
									</table>
								</td>
							<tr>
								<td style="height:20px;">&nbsp;</td>
							</tr>
							<tr>
								<td style="text-align:center;">
									<p style="font-size:14px; color:rgba(69, 80, 86, 0.7411764705882353); line-height:18px; margin:0 0 0;">&copy; <strong>Watchlistfy</strong></p>
								</td>
							</tr>
							<tr>
								<td style="height:80px;">&nbsp;</td>
							</tr>
						</table>
					</td>
				</tr>
			</table>
		</body>
		</html>`,
	)
	err := e.Send("smtp.gmail.com:587", smtp.PlainAuth("", os.Getenv("FROM_MAIL"), os.Getenv("FROM_MAIL_PASSWORD"), "smtp.gmail.com"))
	if err != nil {
		return err
	}

	return nil
}

func SendPasswordChangedEmail(content, mail string) error {
	e := email.NewEmail()
	e.From = "Watchlistfy <" + os.Getenv("FROM_MAIL") + ">"
	e.To = []string{mail}
	e.Subject = "Password Reset"
	e.HTML = []byte(
		`<!doctype html>
		<html lang="en-US">
		<head>
			<meta content="text/html; charset=utf-8" http-equiv="Content-Type" />
			<title>New Password</title>
			<meta name="description" content="New password due to password request.">
			<style type="text/css">
				a:hover {text-decoration: underline !important;}
			</style>
		</head>
		<body marginheight="0" topmargin="0" marginwidth="0" style="margin: 0px; background-color: #f2f3f8;" leftmargin="0">
			<table cellspacing="0" border="0" cellpadding="0" width="100%" bgcolor="#f2f3f8"
				style="@import url(https://fonts.googleapis.com/css?family=Rubik:300,400,500,700|Open+Sans:300,400,600,700); font-family: 'Open Sans', sans-serif;">
				<tr>
					<td>
						<table style="background-color: #f2f3f8; max-width:670px;  margin:0 auto;" width="100%" border="0"
							align="center" cellpadding="0" cellspacing="0">
							<tr>
								<td style="height:80px;">&nbsp;</td>
							</tr>
							<tr>
								<td style="text-align:center;">
									<img width="125" src="https://user-images.githubusercontent.com/25686023/269638549-fc57304b-bc39-4e66-ae26-7cfc59408a21.png" title="logo" alt="logo">
								</td>
							</tr>
							<tr>
								<td style="height:20px;">&nbsp;</td>
							</tr>
							<tr>
								<td>
									<table width="95%" border="0" align="center" cellpadding="0" cellspacing="0"
										style="max-width:670px;background:#fff; border-radius:3px; text-align:center;-webkit-box-shadow:0 6px 18px 0 rgba(0,0,0,.06);-moz-box-shadow:0 6px 18px 0 rgba(0,0,0,.06);box-shadow:0 6px 18px 0 rgba(0,0,0,.06);">
										<tr>
											<td style="height:40px;">&nbsp;</td>
										</tr>
										<tr>
											<td style="padding:0 35px;">
												<h1 style="color:#1e1e2d; font-weight:500; margin:0;font-size:32px;font-family:'Rubik',sans-serif;">Your password changed</h1>
												<span
													style="display:inline-block; vertical-align:middle; margin:29px 0 26px; border-bottom:1px solid #cecece; width:100px;"></span>
												<p style="color:red; font-size:16px;line-height:24px; margin:0;">New password: </p>
												<p style="color:black">` + content + `</p>
												<br>
											<p style="font-size:14px">Don't forget to change your password later!</p>
											</td>
										</tr>
										<tr>
											<td style="height:40px;">&nbsp;</td>
										</tr>
									</table>
								</td>
							<tr>
								<td style="height:20px;">&nbsp;</td>
							</tr>
							<tr>
								<td style="text-align:center;">
									<p style="font-size:14px; color:rgba(69, 80, 86, 0.7411764705882353); line-height:18px; margin:0 0 0;">&copy; <strong>Watchlistfy</strong></p>
								</td>
							</tr>
							<tr>
								<td style="height:80px;">&nbsp;</td>
							</tr>
						</table>
					</td>
				</tr>
			</table>
		</body>
		</html>
		`,
	)
	err := e.Send("smtp.gmail.com:587", smtp.PlainAuth("", os.Getenv("FROM_MAIL"), os.Getenv("FROM_MAIL_PASSWORD"), "smtp.gmail.com"))
	if err != nil {
		return err
	}

	return nil
}
