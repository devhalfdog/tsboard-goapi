package templates

var VerificationBody string = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verification Code</title>
    <style>
        body {
            font-family: 'Roboto', Arial, sans-serif;
            margin: 0;
            padding: 0;
        }
        .email-container {
            max-width: 450px;
            margin: 0 auto;
            background-color: #ECEFF1;
            padding: 25px;
            border-radius: 20px;
        }
        .header {
            text-align: center;
            padding-bottom: 20px;
        }
        .header h1 {
            color: #263238;
        }
        .content {
            text-align: center;
            font-size: 16px;
            color: #333333;
        }
        .code {
            font-size: 48px;
            font-weight: bold;
            letter-spacing: 8px;
            color: #263238;
            margin: 20px 0;
        }
        .footer {
            text-align: center;
            font-size: 12px;
            color: #888888;
            margin-top: 20px;
        }
        .footer a {
            color: #546E7A;
            text-decoration: none;
        }
        .left {
            text-align: left;
        }
    </style>
</head>
<body>
    <div class="email-container">
        <div class="header">
            <h1>Verify Your Email Address</h1>
        </div>
        <div class="content">
            <h2>Hello {{Name}},</h2>
            <p class="left">Thank you for signing up. To complete your registration, please use the following verification code:</p>
            <div class="code">{{Code}}</div>
            <p class="left">This code will expire in 10 minutes. If you did not request this, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>If you have any questions, contact us at <a href="mailto:{{From}}">{{From}}</a> ⎯ <a href="http://{{Host}}">{{Host}}</a></p>
        </div>
    </div>
</body>
</html>
`
